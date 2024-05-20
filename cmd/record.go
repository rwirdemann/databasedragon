package cmd

import (
	"fmt"
	"github.com/rwirdemann/datafrog/internal/datafrog"
	"github.com/rwirdemann/datafrog/internal/datafrog/mysql"
	"github.com/rwirdemann/datafrog/internal/datafrog/record"
	"log"
	"os"

	"github.com/rwirdemann/datafrog/adapter"
	"github.com/spf13/cobra"
)

func init() {
	recordCmd.Flags().String("out", "", "Filename to save recording")
	recordCmd.Flags().Bool("prompt", false, "Wait for key stroke before recording starts")
	_ = recordCmd.MarkFlagRequired("out")
	rootCmd.AddCommand(recordCmd)
}

// close done channel to stop recording loop.
var recordingDone = make(chan struct{})

// read from stopped channel to wait for the recorder to finish
var recordingStopped = make(chan struct{})

var recorder *record.Recorder
var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Starts recording",
	Run: func(cmd *cobra.Command, args []string) {
		out, _ := cmd.Flags().GetString("out")
		c := datafrog.NewConfig("config.json")
		prompt, _ := cmd.Flags().GetBool("prompt")
		if prompt {
			log.Printf("Recording goes to '%s'. Hit enter when you are ready!", out)
			_, _ = fmt.Scanln()
		} else {
			log.Printf("Recording goes to '%s'.", out)
		}

		recordingSink := adapter.NewFileRecordingSink(out)
		databaseLog := createLogAdapter(c)
		t := &adapter.UTCTimer{}
		recorder = record.NewRecorder(c, mysql.Tokenizer{}, databaseLog, recordingSink, t, out, adapter.GoogleUUIDProvider{})
		go checkExit()
		go recorder.Start(recordingDone, recordingStopped)
		<-recordingStopped
	},
}

// Checks if enter was hit to stop recording.
func checkExit() {
	var b = make([]byte, 1)
	l, _ := os.Stdin.Read(b)
	if l > 0 {
		close(recordingDone)
	}
}
