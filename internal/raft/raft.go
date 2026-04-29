package raft

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	heartbeatInterval = 200 * time.Millisecond
	electionMin       = 5000 * time.Millisecond
	electionMax       = 10000 * time.Millisecond
	submitTimeout     = 5 * time.Second
	rpcTimeout        = 1500 * time.Millisecond
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
		id:       id,
		addr:     strings.TrimRight(addr, "/"),
		peers:    normalizePeers(peers),
		role:     Follower,
		term:     0,
		votedFor: "",
		leaderID: "",
		log: []LogEntry{
			{Index: 0, Term: 0, Command: Command{Type: "noop"}},
		},
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   map[string]int{},
		matchIndex:  map[string]int{},
		config: Config{
			Algorithm:       "round_robin",
			ProbeIntervalMs: 1000,
		},
		electionReset: time.Now(),
		stopCh:        make(chan struct{}),
		applyCh:       make(chan LogEntry, 128),
		httpClient: &http.Client{
			Timeout: rpcTimeout,
		},
		stateFile: stateFile,
	}

	n.loadState()
	return n
}

func normalizePeers(peers []string) []string {
	out := make([]string, 0, len(peers))

	for _, p := range peers {
		p = strings.TrimSpace(p)
		p = strings.TrimRight(p, "/")

		if p == "" {
			continue
		}

		out = append(out, p)
	}

	return out
}

func (n *Node) Start() {
	n.logf("node started role=%s term=%d addr=%s peers=%v", n.role, n.term, n.addr, n.peers)
	go n.runElectionTimer()
	go n.runHeartbeatTimer()
	go n.applyLoop()
}

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

	entry := LogEntry{
		Index:   len(n.log),
		Term:    n.term,
		Command: cmd,
	}

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
		role := n.role
		n.mu.Unlock()

		if committed {
			n.logf("submit committed index=%d term=%d", entry.Index, term)
			return SubmitReply{
				Accepted: true,
				Leader:   n.id,
				Index:    entry.Index,
				Term:     term,
			}, nil
		}

		if role != Leader {
			return SubmitReply{
				Accepted: false,
				Leader:   "",
				Index:    entry.Index,
				Term:     term,
			}, errNotLeader
		}

		time.Sleep(30 * time.Millisecond)
	}

	n.logf("submit timeout index=%d", entry.Index)

	return SubmitReply{
		Accepted: false,
		Leader:   n.id,
		Index:    entry.Index,
	}, errors.New("commit timeout")
}

func (n *Node) HandleRequestVote(args RequestVoteArgs) RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Print the received message along with the real time
	currentTime := time.Now().Format(time.RFC3339)
	n.logf("Received RequestVote from %s at %s", args.CandidateID, currentTime)
	
	// CRITICAL RAFT RULE: Prevent disruptive servers (asymmetric network partitions).
	// If we have heard from a valid leader within the minimum election timeout,
	// we assume the leader is still alive and ignore all vote requests.
	// This stops isolated nodes (like node 3) from constantly hijacking the cluster.
	// We also ensure that the Leader itself always rejects disruptive votes.
	if (time.Since(n.electionReset) < electionMin || n.role == Leader) && n.role != Candidate {
		n.logf("HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH")
		n.logf("request-hohoyhhyoh rejected from %s (term %d): heard from leader recently", args.CandidateID, args.Term)
		return RequestVoteReply{Term: n.term, VoteGranted: false}
	}

	if args.Term < n.term {
		return RequestVoteReply{Term: n.term, VoteGranted: false}
	}

	if args.Term > n.term {
		n.term = args.Term
		n.role = Follower
		n.votedFor = ""
		n.leaderID = ""
		n.electionReset = time.Now()
		n.logf("request-vote saw higher term=%d from=%s", args.Term, args.CandidateID)
		n.persistLocked()
	}

	lastIndex, lastTerm := n.lastLogInfo()

	upToDate := args.LastLogTerm > lastTerm ||
		(args.LastLogTerm == lastTerm && args.LastLogIndex >= lastIndex)

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

	/*
		CRITICAL FIX:

		A node must never process its own AppendEntries as an external heartbeat.

		In multi-laptop or Docker setups, the same node may be reachable through
		different addresses:
			- http://localhost:8084
			- http://127.0.0.1:8084
			- http://192.168.x.x:8084
			- http://node4:8084

		So even if replicateAll tries to skip self by address, a bad address config
		can still send the request back to the same node. LeaderID is the stronger
		identity check.
	*/
	if args.LeaderID == n.id {
		if args.Term > n.term {
			n.term = args.Term
			n.votedFor = ""
			n.persistLocked()
		}

		return AppendEntriesReply{
			Term:    n.term,
			Success: true,
		}
	}

	if args.Term < n.term {
		return AppendEntriesReply{Term: n.term, Success: false}
	}

	if args.Term > n.term {
		n.term = args.Term
		n.role = Follower
		n.votedFor = ""
		n.leaderID = args.LeaderID
		n.electionReset = time.Now()
		n.logf("hiii append-entries saw higher term=%d leader=%s", args.Term, args.LeaderID)
		n.persistLocked()
	} else {
		/*
			Same term AppendEntries from a different node.

			In standard Raft, if a valid leader sends AppendEntries in the same term,
			a candidate/follower accepts it and resets election timeout.

			If this node is also Leader in the same term, that indicates split-brain
			or bad configuration. We step down for safety and log loudly.
		*/
		if n.role == Leader {
			n.logf(
				"WARNING: same-term AppendEntries from another leader=%s while I am leader term=%d; stepping down",
				args.LeaderID,
				n.term,
			)
		}

		n.role = Follower
		n.leaderID = args.LeaderID
		n.electionReset = time.Now()
	}

	if args.PrevLogIndex >= len(n.log) {
		n.logf(
			"append-entries reject: prev index out of range prev=%d log_len=%d",
			args.PrevLogIndex,
			len(n.log),
		)
		return AppendEntriesReply{Term: n.term, Success: false}
	}

	if args.PrevLogIndex >= 0 && n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		n.logf(
			"append-entries reject: prev term mismatch prev=%d local_term=%d leader_term=%d",
			args.PrevLogIndex,
			n.log[args.PrevLogIndex].Term,
			args.PrevLogTerm,
		)
		return AppendEntriesReply{Term: n.term, Success: false}
	}

	index := args.PrevLogIndex + 1

	for i, entry := range args.Entries {
		target := index + i

		if target < len(n.log) {
			if n.log[target].Term != entry.Term {
				n.log = n.log[:target]
				n.log = append(n.log, entry)
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
				n.logf("why is this guy here")
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
			isLeader := n.role == Leader
			if isLeader {
				// Keep the leader's electionReset fresh so it correctly evaluates the
				// "prevent disruptive servers" rule in HandleRequestVote.
				n.electionReset = time.Now()
			}
			n.mu.Unlock()

			if isLeader {
				n.replicateAll()
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

	var votesMu sync.Mutex
	var wg sync.WaitGroup

	for _, peer := range peers {
		if n.isSelfPeer(peer) {
			n.logf("skipping self during election peer=%s addr=%s", peer, n.addr)
			continue
		}

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
				n.logf("request-vote failed peer=%s err=%v", peer, err)
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			if reply.Term > n.term {
				n.term = reply.Term
				n.role = Follower
				n.votedFor = ""
				n.leaderID = ""
				n.electionReset = time.Now()
				n.logf("request-vote reply had higher term=%d peer=%s", reply.Term, peer)
				n.persistLocked()
				return
			}

			if n.role != Candidate || n.term != term {
				return
			}

			if reply.VoteGranted {
				votesMu.Lock()
				votes++
				votesMu.Unlock()
			}
		}()
	}

	wg.Wait()

	n.mu.Lock()
	defer n.mu.Unlock()

	won := votes >= n.majority()

	if won {
		if n.role == Candidate && n.term == term {
			n.becomeLeader()
			go n.replicateAll()
		}
	} else {
		n.logf("election lost term=%d votes=%d required=%d", term, votes, n.majority())
	}
}

func (n *Node) becomeLeader() {
	n.role = Leader
	n.leaderID = n.id

	n.logf("became leader term=%d majority=%d", n.term, n.majority())

	lastIndex := len(n.log)

	for _, peer := range n.peers {
		if n.isSelfPeer(peer) {
			n.logf("skipping self while initializing leader replication peer=%s addr=%s", peer, n.addr)
			continue
		}

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

	for _, peer := range peers {
		if n.isSelfPeer(peer) {
			n.logf("skipping self during replication peer=%s addr=%s", peer, n.addr)
			continue
		}

		peer := peer
		go n.replicateToPeer(peer)
	}
}

func (n *Node) replicateToPeer(peer string) {
	n.mu.Lock()

	if n.role != Leader {
		n.mu.Unlock()
		return
	}

	if n.isSelfPeer(peer) {
		n.logf("refusing to replicate to self peer=%s addr=%s", peer, n.addr)
		n.mu.Unlock()
		return
	}

	next := n.nextIndex[peer]

	if next <= 0 {
		next = len(n.log)
		n.nextIndex[peer] = next
	}

	if next > len(n.log) {
		next = len(n.log)
		n.nextIndex[peer] = next
	}

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
		n.logf("append-entries failed peer=%s err=%v", peer, err)
		return
	}

	n.mu.Lock()

	if reply.Term > n.term {
		n.term = reply.Term
		n.role = Follower
		n.votedFor = ""
		n.leaderID = ""
		n.electionReset = time.Now()
		n.logf("append-entries reply had higher term=%d peer=%s", reply.Term, peer)
		n.persistLocked()
		n.mu.Unlock()
		return
	}

	if n.role != Leader {
		n.mu.Unlock()
		return
	}

	shouldUpdateCommit := false

	if reply.Success {
		n.matchIndex[peer] = args.PrevLogIndex + len(args.Entries)
		n.nextIndex[peer] = n.matchIndex[peer] + 1
		shouldUpdateCommit = true
	} else {
		n.logf("replication failed peer=%s next_index=%d", peer, n.nextIndex[peer])

		if n.nextIndex[peer] > 1 {
			n.nextIndex[peer]--
		}
	}

	n.mu.Unlock()

	if shouldUpdateCommit {
		n.updateCommitIndex()
	}
}

func (n *Node) updateCommitIndex() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.role != Leader {
		return
	}

	for i := n.commitIndex + 1; i < len(n.log); i++ {
		count := 1 // leader itself

		for _, peer := range n.peers {
			if n.isSelfPeer(peer) {
				continue
			}

			if n.matchIndex[peer] >= i {
				count++
			}
		}

		if count >= n.majority() && n.log[i].Term == n.term {
			oldCommit := n.commitIndex
			n.commitIndex = i
			n.persistLocked()
			n.logf("leader commit advanced old=%d new=%d count=%d required=%d", oldCommit, n.commitIndex, count, n.majority())
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

		n.mu.Lock()
		n.config = cfg
		obs := append([]func(Config){}, n.configObservers...)
		n.persistLocked()
		n.mu.Unlock()

		n.logf(
			"applied config algorithm=%s probe_interval_ms=%d",
			cfg.Algorithm,
			cfg.ProbeIntervalMs,
		)

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
		st.Log = []LogEntry{
			{Index: 0, Term: 0, Command: Command{Type: "noop"}},
		}
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
	/*
		Your current design says peers includes every configured node,
		including self.

		So:
			3 nodes -> majority 2
			5 nodes -> majority 3
			1 node  -> majority 1
	*/
	return len(n.peers)/2 + 1
}

func (n *Node) isSelfPeer(peer string) bool {
	return sameRaftAddress(peer, n.addr)
}

func sameRaftAddress(a, b string) bool {
	a = strings.TrimRight(strings.TrimSpace(a), "/")
	b = strings.TrimRight(strings.TrimSpace(b), "/")

	if a == b {
		return true
	}

	ua, errA := url.Parse(a)
	ub, errB := url.Parse(b)

	if errA != nil || errB != nil {
		return a == b
	}

	hostA, portA, errA := net.SplitHostPort(ua.Host)
	hostB, portB, errB := net.SplitHostPort(ub.Host)

	if errA != nil || errB != nil {
		return a == b
	}

	if portA != portB {
		return false
	}

	hostA = normalizeHost(hostA)
	hostB = normalizeHost(hostB)

	if hostA == hostB {
		return true
	}

	localA := isLocalHost(hostA)
	localB := isLocalHost(hostB)

	return localA && localB
}

func normalizeHost(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.Trim(h, "[]")
	return h
}

func isLocalHost(h string) bool {
	return h == "localhost" || h == "127.0.0.1" || h == "::1" || h == "0.0.0.0"
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

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		bytes.NewBuffer(b),
	)
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