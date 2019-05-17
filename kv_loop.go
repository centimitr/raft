package raft

func (kv *KV) loop() {
	for {
		select {
		case <-kv.shutdown:
			debug("shutdown: apply")
			return
		case msg := <-kv.applyCh:
			// use snapshot
			if msg.UseSnapshot {
				kv.mu.Lock()
				kv.LoadFromData(msg.Snapshot)
				kv.Store(msg.Index)
				kv.mu.Unlock()
				continue
			}

			// check message
			if msg.Command == nil || msg.Index <= kv.snapshotIndex {
				continue
			}

			cmd := msg.Command.(cmd)
			kv.mu.Lock()
			switch cmd.Type {
			case "Get":
			case "Put":
				kv.m[cmd.Key] = cmd.Value
			case "Append":
				kv.m[cmd.Key] += cmd.Value
			default:
				panic("kv: invalid op")
			}

			// check if need snapshot
			if kv.needSnapshot() {
				debug("need snapshot:", kv.snapshotThreshold, kv.store.StateSize())
				kv.Store(msg.Index)
				kv.raft.DidSnapshot(msg.Index)
			}

			// notify channel
			ch, ok := kv.notify[msg.Index]
			if ok && ch != nil {
				close(ch)
				delete(kv.notify, msg.Index)
			}

			kv.mu.Unlock()
		}
	}
}
