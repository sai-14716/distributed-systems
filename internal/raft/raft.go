package raft

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
	"fmt"
	"log"
    "math/rand"
    "net/http"
	"os"
	"path/filepath"
    "sync"
    "time"
)

const (
	heartbeatInterval = 100 * time.Millisecond
	electionMin      = 300 * time.Millisecond
	electionMax      = 600 * time.Millisecond
	submitTimeout    = 3 * time.Second
)

var errNotLeader = errors.New("not leader")

type Node struct {
	mu sync.Mutex

	id    string
	addr  string
	peers []string

	role     Role
	term     int
	votedFor string
	leaderID string

	log         []LogEntry
	commitIndex int
	lastApplied int

	nextIndex  map[string]int
	matchIndex map[string]int

	config Config

	configObservers []func(Config)

	electionReset time.Time

	stopCh  chan struct{}
	applyCh chan LogEntry

	httpClient *http.Client

	stateFile string
}

type persistedState struct {
	Term        int        `json:"term"`
	VotedFor    string     `json:"voted_for"`
	Log         []LogEntry `json:"log"`
	CommitIndex int        `json:"commit_index"`
	LastApplied int        `json:"last_applied"`
	Config      Config     `json:"config"`
}

func NewNode(id, addr string, peers []string, stateFile string) *Node {
	n := &Node{
		id:      id,
		addr:    addr,
		peers:   peers,
		role:    Follower,
		term:    0,
		votedFor: "",
		leaderID: "",
		log: []LogEntry{{Index: 0, Term: 0, Command: Command{Type: "noop"}}},
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   map[string]int{},
		matchIndex:  map[string]int{},
		config: Config{
			Algorithm:       "round_robin",
			ProbeIntervalMs: 1000,
			HealthThreshold: 0.8,
		},
		electionReset: time.Now(),
		stopCh:        make(chan struct{}),
		applyCh:       make(chan LogEntry, 128),
		httpClient: &http.Client{
			Timeout: 800 * time.Millisecond,
		},
		stateFile: stateFile,
	}
	n.loadState()
	return n
}

func (n *Node) Start() {
	n.logf("node started role=%s term=%d", n.role, n.term)
	go n.runElectionTimer()
	go n.runHeartbeatTimer()
	go n.applyLoop()
}

// OnConfigApplied registers a callback invoked after config is applied from the log.
// The callback is invoked in the apply loop goroutine.
func (n *Node) OnConfigApplied(fn func(Config)) {
	if fn == nil {
		return
	}
	n.mu.Lock()
	n.configObservers = append(n.configObservers, fn)
	n.mu.Unlock()
}

func (n *Node) Stop() {
	close(n.stopCh)
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Addr() string {
	return n.addr
}

func (n *Node) State() StateView {
	n.mu.Lock()
	defer n.mu.Unlock()
	return StateView{
		ID:          n.id,
		Role:        n.role,
		Term:        n.term,
		CommitIndex: n.commitIndex,
		LastApplied: n.lastApplied,
		LeaderID:    n.leaderID,
		Config:      n.config,
		LogLen:      len(n.log),
	}
}

func (n *Node) Submit(cmd Command) (SubmitReply, error) {
	n.mu.Lock()
	if n.role != Leader {
		leader := n.leaderID
		n.logf("submit rejected: not leader leader=%s cmd=%s", leader, cmd.Type)
		n.mu.Unlock()
		return SubmitReply{Accepted: false, Leader: leader}, errNotLeader
	}
	entry := LogEntry{Index: len(n.log), Term: n.term, Command: cmd}
	n.log = append(n.log, entry)
	n.persistLocked()
	n.logf("submit accepted: appended index=%d term=%d cmd=%s", entry.Index, entry.Term, cmd.Type)
	n.mu.Unlock()

	n.replicateAll()

	deadline := time.Now().Add(submitTimeout)
	for time.Now().Before(deadline) {
		n.mu.Lock()
		committed := n.commitIndex >= entry.Index
		term := n.term
		n.mu.Unlock()
		if committed {
			n.logf("submit committed index=%d term=%d", entry.Index, term)
			return SubmitReply{Accepted: true, Leader: n.id, Index: entry.Index, Term: term}, nil
		}
		time.Sleep(30 * time.Millisecond)
	}
	n.logf("submit timeout index=%d", entry.Index)
	return SubmitReply{Accepted: false, Leader: n.id, Index: entry.Index}, errors.New("commit timeout")
}

func (n *Node) HandleRequestVote(args RequestVoteArgs) RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.term {
		return RequestVoteReply{Term: n.term, VoteGranted: false}
	}

	if args.Term > n.term {
		n.term = args.Term
		n.role = Follower
		n.votedFor = ""
		n.logf("request-vote saw higher term=%d from=%s", args.Term, args.CandidateID)
		n.persistLocked()
	}

	lastIndex, lastTerm := n.lastLogInfo()
	upToDate := args.LastLogTerm > lastTerm || (args.LastLogTerm == lastTerm && args.LastLogIndex >= lastIndex)

	if (n.votedFor == "" || n.votedFor == args.CandidateID) && upToDate {
		n.votedFor = args.CandidateID
		n.electionReset = time.Now()
		n.logf("vote granted to=%s term=%d", args.CandidateID, n.term)
		n.persistLocked()
		return RequestVoteReply{Term: n.term, VoteGranted: true}
	}

	return RequestVoteReply{Term: n.term, VoteGranted: false}
}

func (n *Node) HandleAppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term < n.term {
		return AppendEntriesReply{Term: n.term, Success: false}
	}

	if args.Term > n.term {
		n.term = args.Term
		n.votedFor = ""
		n.logf("append-entries saw higher term=%d leader=%s", args.Term, args.LeaderID)
		n.persistLocked()
	}

	n.role = Follower
	n.leaderID = args.LeaderID
	n.electionReset = time.Now()

	if args.PrevLogIndex >= len(n.log) {
		n.logf("append-entries reject: prev index out of range prev=%d log_len=%d", args.PrevLogIndex, len(n.log))
		return AppendEntriesReply{Term: n.term, Success: false}
	}
	if args.PrevLogIndex >= 0 && n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		n.logf("append-entries reject: prev term mismatch prev=%d local_term=%d leader_term=%d", args.PrevLogIndex, n.log[args.PrevLogIndex].Term, args.PrevLogTerm)
		return AppendEntriesReply{Term: n.term, Success: false}
	}

	// Append new entries, overwriting conflicts
	index := args.PrevLogIndex + 1
	for i, entry := range args.Entries {
		if index+i < len(n.log) {
			if n.log[index+i].Term != entry.Term {
				n.log = n.log[:index+i]
				n.log = append(n.log, entry)
			} else {
				// entry already matches
			}
		} else {
			n.log = append(n.log, entry)
		}
	}
	if len(args.Entries) > 0 {
		n.logf("append-entries applied entries=%d leader=%s", len(args.Entries), args.LeaderID)
	}

	if args.LeaderCommit > n.commitIndex {
		oldCommit := n.commitIndex
		lastIndex := len(n.log) - 1
		if args.LeaderCommit < lastIndex {
			n.commitIndex = args.LeaderCommit
		} else {
			n.commitIndex = lastIndex
		}
		if n.commitIndex != oldCommit {
			n.logf("commit index updated old=%d new=%d leader=%s", oldCommit, n.commitIndex, args.LeaderID)
		}
	}
	n.persistLocked()

	return AppendEntriesReply{Term: n.term, Success: true}
}

func (n *Node) runElectionTimer() {
	for {
		timeout := randomElectionTimeout()
		select {
		case <-n.stopCh:
			return
		case <-time.After(timeout):
			n.mu.Lock()
			if n.role == Leader {
				n.mu.Unlock()
				continue
			}
			if time.Since(n.electionReset) < timeout {
				n.mu.Unlock()
				continue
			}
			n.mu.Unlock()
			n.startElection()
		}
	}
}

func (n *Node) runHeartbeatTimer() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			n.mu.Lock()
			if n.role == Leader {
				n.mu.Unlock()
				n.replicateAll()
			} else {
				n.mu.Unlock()
			}
		}
	}
}

func (n *Node) startElection() {
	n.mu.Lock()
	n.role = Candidate
	n.term++
	n.votedFor = n.id
	n.leaderID = ""
	n.electionReset = time.Now()
	n.logf("election started term=%d", n.term)
	n.persistLocked()

	votes := 1
	term := n.term
	candidateID := n.id
	lastIndex, lastTerm := n.lastLogInfo()
	peers := append([]string{}, n.peers...)
	n.mu.Unlock()

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, peer := range peers {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			args := RequestVoteArgs{
				Term:         term,
				CandidateID:  candidateID,
				LastLogIndex: lastIndex,
				LastLogTerm:  lastTerm,
			}
			reply, err := n.sendRequestVote(peer, args)
			if err != nil {
				return
			}
			n.mu.Lock()
			defer n.mu.Unlock()
			if reply.Term > n.term {
				n.term = reply.Term
				n.role = Follower
				n.votedFor = ""
				return
			}
			if n.role != Candidate || n.term != term {
				return
			}
			if reply.VoteGranted {
				mu.Lock()
				votes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	mu.Lock()
	won := votes >= n.majority()
	mu.Unlock()
	if won {
		n.mu.Lock()
		if n.role == Candidate && n.term == term {
			n.becomeLeader()
		}
		n.mu.Unlock()
	} else {
		n.logf("election lost term=%d votes=%d", term, votes)
	}
}

func (n *Node) becomeLeader() {
	n.role = Leader
	n.leaderID = n.id
	n.logf("became leader term=%d", n.term)
	lastIndex := len(n.log)
	for _, peer := range n.peers {
		n.nextIndex[peer] = lastIndex
		n.matchIndex[peer] = 0
	}
}

func (n *Node) replicateAll() {
	n.mu.Lock()
	if n.role != Leader {
		n.mu.Unlock()
		return
	}
	peers := append([]string{}, n.peers...)
	n.mu.Unlock()

	var wg sync.WaitGroup
	for _, peer := range peers {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.replicateToPeer(peer)
		}()
	}
	wg.Wait()

	n.updateCommitIndex()
}

func (n *Node) replicateToPeer(peer string) {
	n.mu.Lock()
	next := n.nextIndex[peer]
	prevIndex := next - 1
	prevTerm := 0
	if prevIndex >= 0 && prevIndex < len(n.log) {
		prevTerm = n.log[prevIndex].Term
	}
	entries := make([]LogEntry, len(n.log[next:]))
	copy(entries, n.log[next:])
	args := AppendEntriesArgs{
		Term:         n.term,
		LeaderID:     n.id,
		PrevLogIndex: prevIndex,
		PrevLogTerm:  prevTerm,
		Entries:      entries,
		LeaderCommit: n.commitIndex,
	}
	n.mu.Unlock()

	reply, err := n.sendAppendEntries(peer, args)
	if err != nil {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	if reply.Term > n.term {
		n.term = reply.Term
		n.role = Follower
		n.votedFor = ""
		n.persistLocked()
		return
	}
	if n.role != Leader {
		return
	}
	if reply.Success {
		n.matchIndex[peer] = args.PrevLogIndex + len(args.Entries)
		n.nextIndex[peer] = n.matchIndex[peer] + 1
	} else {
		n.logf("replication failed peer=%s next_index=%d", peer, n.nextIndex[peer])
		if n.nextIndex[peer] > 1 {
			n.nextIndex[peer]--
		}
	}
}

func (n *Node) updateCommitIndex() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.role != Leader {
		return
	}
	for i := n.commitIndex + 1; i < len(n.log); i++ {
		count := 1
		for _, peer := range n.peers {
			if n.matchIndex[peer] >= i {
				count++
			}
		}
		if count >= n.majority() && n.log[i].Term == n.term {
			oldCommit := n.commitIndex
			n.commitIndex = i
			n.persistLocked()
			n.logf("leader commit advanced old=%d new=%d", oldCommit, n.commitIndex)
		}
	}
}

func (n *Node) applyLoop() {
	for {
		select {
		case <-n.stopCh:
			return
		default:
			n.mu.Lock()
			if n.lastApplied < n.commitIndex {
				n.lastApplied++
				entry := n.log[n.lastApplied]
				n.persistLocked()
				n.mu.Unlock()
				n.applyEntry(entry)
			} else {
				n.mu.Unlock()
				time.Sleep(20 * time.Millisecond)
			}
		}
	}
}

func (n *Node) applyEntry(entry LogEntry) {
	switch entry.Command.Type {
	case "set_config":
		cfg := n.config
		if v, ok := entry.Command.Data["algorithm"].(string); ok {
			cfg.Algorithm = v
		}
		if v, ok := entry.Command.Data["probe_interval_ms"].(float64); ok {
			cfg.ProbeIntervalMs = int(v)
		}
		if v, ok := entry.Command.Data["health_threshold"].(float64); ok {
			cfg.HealthThreshold = v
		}
		n.mu.Lock()
		n.config = cfg
		obs := append([]func(Config){}, n.configObservers...)
		n.persistLocked()
		n.mu.Unlock()
		n.logf("applied config algorithm=%s probe_interval_ms=%d health_threshold=%.3f", cfg.Algorithm, cfg.ProbeIntervalMs, cfg.HealthThreshold)
		for _, fn := range obs {
			fn(cfg)
		}
	}
}

func (n *Node) logf(format string, args ...any) {
	log.Printf("node=%s "+format, append([]any{n.id}, args...)...)
}

func (n *Node) loadState() {
	if n.stateFile == "" {
		return
	}
	b, err := os.ReadFile(n.stateFile)
	if err != nil {
		return
	}
	var st persistedState
	if err := json.Unmarshal(b, &st); err != nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(st.Log) == 0 {
		st.Log = []LogEntry{{Index: 0, Term: 0, Command: Command{Type: "noop"}}}
	}
	n.term = st.Term
	n.votedFor = st.VotedFor
	n.log = st.Log
	n.commitIndex = st.CommitIndex
	n.lastApplied = st.LastApplied
	n.config = st.Config
	if n.lastApplied > n.commitIndex {
		n.lastApplied = n.commitIndex
	}
	if n.commitIndex >= len(n.log) {
		n.commitIndex = len(n.log) - 1
	}
}

func (n *Node) persistLocked() {
	if n.stateFile == "" {
		return
	}
	st := persistedState{
		Term:        n.term,
		VotedFor:    n.votedFor,
		Log:         append([]LogEntry(nil), n.log...),
		CommitIndex: n.commitIndex,
		LastApplied: n.lastApplied,
		Config:      n.config,
	}
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(n.stateFile), 0o755); err != nil {
		return
	}
	tmpPath := fmt.Sprintf("%s.tmp", n.stateFile)
	if err := os.WriteFile(tmpPath, b, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmpPath, n.stateFile)
}

func (n *Node) lastLogInfo() (int, int) {
	lastIndex := len(n.log) - 1
	lastTerm := n.log[lastIndex].Term
	return lastIndex, lastTerm
}

func (n *Node) majority() int {
	return (len(n.peers)+1)/2 + 1
}

func randomElectionTimeout() time.Duration {
	return electionMin + time.Duration(rand.Int63n(int64(electionMax-electionMin)))
}

func (n *Node) sendRequestVote(peer string, args RequestVoteArgs) (RequestVoteReply, error) {
	var reply RequestVoteReply
	if err := n.postJSON(peer+"/raft/request-vote", args, &reply); err != nil {
		return reply, err
	}
	return reply, nil
}

func (n *Node) sendAppendEntries(peer string, args AppendEntriesArgs) (AppendEntriesReply, error) {
	var reply AppendEntriesReply
	if err := n.postJSON(peer+"/raft/append-entries", args, &reply); err != nil {
		return reply, err
	}
	return reply, nil
}

func (n *Node) postJSON(url string, reqBody any, respBody any) error {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("non-2xx response")
	}
	if respBody != nil {
		return json.NewDecoder(resp.Body).Decode(respBody)
	}
	return nil
}
