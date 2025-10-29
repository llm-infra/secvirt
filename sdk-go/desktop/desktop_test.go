package desktop

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
	"trpc.group/trpc-go/trpc-a2a-go/client"
	"trpc.group/trpc-go/trpc-a2a-go/log"
	"trpc.group/trpc-go/trpc-a2a-go/protocol"
)

func TestSandboxRun(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)
	defer sbx.DestroySandbox(t.Context())

	cli, err := sbx.NewA2AClient(t.Context())
	assert.NoError(t, err)

	msgParams := protocol.SendMessageParams{
		Message: protocol.NewMessage(
			protocol.MessageRoleUser,
			[]protocol.Part{protocol.NewTextPart("hello")},
		),
	}

	streamChan, err := cli.StreamMessage(t.Context(), msgParams)
	if err != nil {
		log.Fatalf("Error starting stream task: %v.", err)
	}

	processStreamEventsWithInteraction(t.Context(), cli, streamChan)
}

func processStreamEventsWithInteraction(ctx context.Context, c *client.A2AClient, streamChan <-chan protocol.StreamingMessageEvent) {
	log.Info("Waiting for streaming updates...")

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			// Context timed out or was cancelled
			log.Infof("Streaming context done: %v", ctx.Err())
			return
		case event, ok := <-streamChan:
			if !ok {
				// Channel closed by the client/server
				log.Info("Stream closed.")
				if ctx.Err() != nil {
					log.Infof("Context error after stream close: %v", ctx.Err())
				}
				return
			}

			// Process the received event
			switch event.Result.GetKind() {
			case protocol.KindMessage:
				msg := event.Result.(*protocol.Message)
				text := extractTextFromMessage(msg)
				log.Infof("┌─────────────────────────────────────────")
				log.Infof("│ MESSAGE RECEIVED")
				log.Infof("│ MessageID: %s", msg.MessageID)
				log.Infof("│ Role: %s", msg.Role)
				log.Infof("│ Text: %s", text)
				log.Infof("└─────────────────────────────────────────")

			case protocol.KindTask:
				task := event.Result.(*protocol.Task)
				log.Infof("┌─────────────────────────────────────────")
				log.Infof("│ TASK RECEIVED")
				log.Infof("│ TaskID: %s", task.ID)
				log.Infof("│ State: %s", task.Status.State)
				if task.Status.Message != nil {
					log.Infof("│ Status Message: %s", *task.Status.Message)
				}
				log.Infof("└─────────────────────────────────────────")

			case protocol.KindTaskStatusUpdate:
				statusUpdate := event.Result.(*protocol.TaskStatusUpdateEvent)
				log.Infof("┌─────────────────────────────────────────")
				log.Infof("│ TASK STATUS UPDATE")
				log.Infof("│ TaskID: %s", statusUpdate.TaskID)
				log.Infof("│ State: %s", statusUpdate.Status.State)
				if statusUpdate.Status.Message != nil {
					log.Infof("│ Status Message: %s", *statusUpdate.Status.Message)
				}
				log.Infof("└─────────────────────────────────────────")

				// Handle input-required state
				if statusUpdate.Status.State == protocol.TaskStateInputRequired {
					log.Info("")
					log.Info("⚠️  Task requires user input!")
					log.Info("📝 Please provide your response:")
					fmt.Print("> ")

					// Read user input
					userInput, err := reader.ReadString('\n')
					if err != nil {
						log.Errorf("Error reading user input: %v", err)
						continue
					}
					userInput = strings.TrimSpace(userInput)

					if userInput == "" {
						log.Warn("Empty input provided, skipping...")
						continue
					}

					log.Infof("Sending user input: %s", userInput)

					// Send the user's response as a new message
					msgParams := protocol.SendMessageParams{
						Message: protocol.NewMessage(
							protocol.MessageRoleUser,
							[]protocol.Part{protocol.NewTextPart(userInput)},
						),
					}

					// Continue the conversation with streaming
					newStreamChan, err := c.StreamMessage(ctx, msgParams)
					if err != nil {
						log.Errorf("Error sending user response: %v", err)
						continue
					}

					log.Info("✅ User input sent, continuing to listen for updates...")
					// Replace the stream channel with the new one
					streamChan = newStreamChan
				}

			case protocol.KindTaskArtifactUpdate:
				artifact := event.Result.(*protocol.TaskArtifactUpdateEvent)
				log.Infof("┌─────────────────────────────────────────")
				log.Infof("│ ARTIFACT UPDATE")
				log.Infof("│ TaskID: %s", artifact.TaskID)
				log.Infof("│ ArtifactID: %s", artifact.Artifact.ArtifactID)
				if artifact.LastChunk != nil {
					log.Infof("│ LastChunk: %t", *artifact.LastChunk)
				}
				log.Infof("│ Content:")
				for i, part := range artifact.Artifact.Parts {
					if textPart, ok := part.(*protocol.TextPart); ok {
						log.Infof("│   Part %d: %s", i+1, textPart.Text)
					} else {
						log.Infof("│   Part %d: [%T]", i+1, part)
					}
				}
				log.Infof("└─────────────────────────────────────────")

				// For artifact updates, we note it's the final artifact,
				// but we don't exit yet - per A2A spec, we should wait for the final status update
				if artifact.LastChunk != nil && *artifact.LastChunk {
					log.Info("✅ Received final artifact update, waiting for final status.")
				}

			default:
				log.Infof("┌─────────────────────────────────────────")
				log.Infof("│ UNKNOWN EVENT")
				log.Infof("│ Type: %T", event.Result)
				log.Infof("│ Content: %+v", event.Result)
				log.Infof("└─────────────────────────────────────────")
			}
		}
	}
}

func extractTextFromMessage(msg *protocol.Message) string {
	var text string
	for _, part := range msg.Parts {
		if textPart, ok := part.(protocol.TextPart); ok {
			text += textPart.Text
		}
	}
	return text
}
