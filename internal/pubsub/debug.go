package pubsub

// ConnectionSnapshot is a serializable debug view of one PubSub connection.
type ConnectionSnapshot struct {
	Index         int      `json:"index"`
	IsConnected   bool     `json:"is_connected"`
	TopicCount    int      `json:"topic_count"`
	PendingTopics int      `json:"pending_topics"`
	LastMessageID string   `json:"last_message_id,omitempty"`
	Topics        []string `json:"topics"`
}

// Snapshot returns a serializable debug snapshot of the pool state.
func (p *Pool) Snapshot() []ConnectionSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()

	snapshots := make([]ConnectionSnapshot, 0, len(p.conns))
	for _, conn := range p.conns {
		snapshots = append(snapshots, conn.snapshot())
	}

	return snapshots
}

func (c *Connection) snapshot() ConnectionSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	topics := make([]string, 0, len(c.topics))
	for _, topic := range c.topics {
		topics = append(topics, topic.String())
	}

	return ConnectionSnapshot{
		Index:         c.index,
		IsConnected:   c.isConnected,
		TopicCount:    len(c.topics),
		PendingTopics: len(c.pendingTopics),
		LastMessageID: c.lastMsgIdentifier,
		Topics:        topics,
	}
}
