package raft

type Role string

const (
	Follower  Role = "follower"
	Candidate Role = "candidate"
	Leader    Role = "leader"
)

type LogEntry struct {
	Index   int     `json:"index"`
	Term    int     `json:"term"`
	Command Command `json:"command"`
}

type Command struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type RequestVoteArgs struct {
	Term         int    `json:"term"`
	CandidateID  string `json:"candidate_id"`
	LastLogIndex int    `json:"last_log_index"`
	LastLogTerm  int    `json:"last_log_term"`
}

type RequestVoteReply struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"vote_granted"`
}

type AppendEntriesArgs struct {
	Term         int        `json:"term"`
	LeaderID     string     `json:"leader_id"`
	PrevLogIndex int        `json:"prev_log_index"`
	PrevLogTerm  int        `json:"prev_log_term"`
	Entries      []LogEntry `json:"entries"`
	LeaderCommit int        `json:"leader_commit"`
}

type AppendEntriesReply struct {
	Term    int  `json:"term"`
	Success bool `json:"success"`
}

type SubmitReply struct {
	Accepted bool   `json:"accepted"`
	Leader   string `json:"leader"`
	Index    int    `json:"index"`
	Term     int    `json:"term"`
}

type StateView struct {
	ID          string `json:"id"`
	Role        Role   `json:"role"`
	Term        int    `json:"term"`
	CommitIndex int    `json:"commit_index"`
	LastApplied int    `json:"last_applied"`
	LeaderID    string `json:"leader_id"`
	Config      Config `json:"config"`
	LogLen      int    `json:"log_len"`
}

type Config struct {
	Algorithm       string `json:"algorithm"`
	ProbeIntervalMs int    `json:"probe_interval_ms"`
}
