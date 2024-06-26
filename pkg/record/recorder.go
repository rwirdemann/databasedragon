package record

import (
	"github.com/rwirdemann/datafrog/pkg/df"
	log "github.com/sirupsen/logrus"
)

// A Recorder monitors a channel log and records all statements that match one of
// the patterns specified in the channels pattern list. The recorded output is
// written back TestRepository.
type Recorder struct {
	channel        df.Channel
	tokenizer      df.Tokenizer
	log            df.Log
	timer          df.Timer
	testname       string
	uuidProvider   UUIDProvider
	testcase       df.Testcase
	testRepository df.TestRepository
}

// NewRecorder creates a new Recorder.
func NewRecorder(channel df.Channel, tokenizer df.Tokenizer,
	log df.Log, timer df.Timer, testname string,
	uuidProvider UUIDProvider, repository df.TestRepository) *Recorder {

	return &Recorder{
		channel:        channel,
		tokenizer:      tokenizer,
		log:            log,
		timer:          timer,
		testname:       testname,
		uuidProvider:   uuidProvider,
		testcase:       df.Testcase{Name: testname},
		testRepository: repository,
	}
}

// Start starts the recording process of channel as endless loop. Every log entry
// that matches one of the patterns specified in the channels pattern list is
// written to the recording sink. Only log entries that fall in the actual
// recording period are considered.
func (r *Recorder) Start(done chan struct{}, stopped chan struct{}) {
	r.timer.Start()
	log.Printf("Recording started at %v...", r.timer.GetStart())

	// tell caller that recording has been finished
	defer close(stopped)

	// called when done channel is closed
	defer func() {
		if err := r.testRepository.Write(r.testname, r.testcase); err != nil {
			log.Fatal(err)
		}
	}()

	// jump to log file end
	if err := r.log.Tail(); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		default:
			line, err := r.log.NextLine(done)
			if err != nil {
				log.Fatal(err)
			}
			ts, err := r.log.Timestamp(line)
			if err != nil {
				continue
			}
			if r.timer.MatchesRecordingPeriod(ts) {
				matches, pattern := df.MatchesPattern(r.channel.Patterns, line)
				if matches {
					tokens := r.tokenizer.Tokenize(line, r.channel.Patterns)
					e := df.Expectation{Uuid: r.uuidProvider.NewString(), Tokens: tokens, IgnoreDiffs: []int{}, Pattern: pattern}
					r.testcase.Expectations = append(r.testcase.Expectations, e)
					log.Printf("new expectation: %s\n", e.Shorten(8))
				}
			}
		// check if the caller (web, cli, ...) has closed the done channel to
		// tell me that recoding has been finished
		case <-done:
			log.Println("recorder: done channel closed")
			return
		}
	}
}

func (r *Recorder) Testcase() df.Testcase {
	return r.testcase
}
