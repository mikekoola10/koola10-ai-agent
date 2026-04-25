package main

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

func startSwarmListeners() {
	if redisClient == nil { return }
	ctx := context.Background()

	// Orchestrator Listener
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:orchestrator")
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil { continue }
			var task SwarmTask
			json.Unmarshal([]byte(msg.Payload), &task)

			// Simple logic: route to current stage
			channel := "tasks:" + task.Stage
			redisClient.Publish(ctx, channel, msg.Payload)
		}
	}()

	// Finder Listener
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:finding")
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil { continue }
			var task SwarmTask
			json.Unmarshal([]byte(msg.Payload), &task)

			log.Printf("Finder picking up task %s", task.TaskID)
			// Mock search logic (would call handleSearch logic)
			task.Results["grants"] = []string{"Grant 1", "Grant 2"}
			task.Stage = "writing"

			b, _ := json.Marshal(task)
			redisClient.Set(ctx, "task:"+task.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm_agent", 0.01, "FinderAgent search")
		}
	}()

	// Writer Listener
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:writing")
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil { continue }
			var task SwarmTask
			json.Unmarshal([]byte(msg.Payload), &task)

			log.Printf("Writer picking up task %s", task.TaskID)
			task.Results["draft"] = "AI Generated Narrative Draft Content"
			task.Stage = "reviewing"

			b, _ := json.Marshal(task)
			redisClient.Set(ctx, "task:"+task.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm_agent", 0.05, "WriterAgent drafting")
		}
	}()

	// Reviewer Listener
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:reviewing")
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil { continue }
			var task SwarmTask
			json.Unmarshal([]byte(msg.Payload), &task)

			log.Printf("Reviewer picking up task %s", task.TaskID)
			eval := EvaluateAction("grant_submit", 0.07) // estimated total cost
			task.Results["review"] = eval

			if eval.Decision == "block" {
				task.Stage = "failed"
			} else {
				task.Stage = "submitting"
			}

			b, _ := json.Marshal(task)
			redisClient.Set(ctx, "task:"+task.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm_agent", 0.01, "ReviewerAgent review")
		}
	}()

	// Submitter Listener
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:submitting")
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil { continue }
			var task SwarmTask
			json.Unmarshal([]byte(msg.Payload), &task)

			log.Printf("Submitter picking up task %s", task.TaskID)
			task.Results["submission"] = "Success! Application ID: SUB123"
			task.Stage = "done"

			b, _ := json.Marshal(task)
			redisClient.Set(ctx, "task:"+task.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm_agent", 0.02, "SubmitterAgent browser action")
		}
	}()
}
