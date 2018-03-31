package lib

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/fatih/color"
)

const (
	Lesson    = "### Lesson "
	Sentences = "### Sentences "
)

type QuestionsAnswers struct {
	questions []string
	answers   []string
}

// Topic represents the list of subsections of the file with the questions
// attached for that section.
type Topic struct {
	list map[string]QuestionsAnswers
}

// TopicParsingParameters is a data structure that helps to parse the lines that
// split the different sections.
type TopicParsingParameters struct {
	// topicAnnounce is the string that is used to announce the section in the
	// csv file. For instance, '### Lesson '
	// The text after this string will be considered as the ID of the topic.
	TopicAnnounce string
	// QaSep is the separator on the line between the question and the answer in
	// the csv file. If this separator is found multiple times on the line, the
	// first one is considered as the separator.
	QaSep string
}

type interrogationMode int

const (
	linear  interrogationMode = iota // will ask questions in the same order as the file
	random                           // will ask questions in a random order
	summary                          // ask to show the list of subsections
)

type InterrogationParameters struct {
	interactive bool
	wait        time.Duration     // Default is to wait 2 seconds
	mode        interrogationMode // Default is random.
	in          io.Reader         // Default is to use io.Stdin. Allows to send command to the engine
	out         io.Writer         // The place where the questions are written to
	subsections string            // the list of selected subsections chosen for the questioning
	limit       int               // Limit is the number of times the list is repeated during interrogation. Default is 10
	reversed    bool              // Requires that questions becomes answers and answers becomes questions
	qachan      chan string       // Experimental. Channel to receive questions and answers
	command     chan string       // Experimental. Channel to receive commands
	publisher   chan string       // Experimental. Channel to publish to the output. This channel collects all that needs to be put to the user.
}

// IsSummaryMode tells if the parameters require to have a summary of the subsections.
func (p InterrogationParameters) IsSummaryMode() bool {
	return p.mode == summary
}

// IsReversedMode tells if the user wants that the left column are now answers and right column(s) are the questions
func (p InterrogationParameters) IsReversedMode() bool {
	return p.reversed
}

// GetOutputStream gets the Writer where questions will be written to.
func (p InterrogationParameters) GetOutputStream() io.Writer {
	return p.out
}

// GetListOfSubsections returns a string array containing all the subsections selected by
// the end user.
func (p InterrogationParameters) GetListOfSubsections() []string {
	if len(p.subsections) == 0 {
		return nil
	}
	return strings.Split(p.subsections, ",")
}

// NewQA builds an empty set of questions/answers.
func NewQA() QuestionsAnswers {
	return QuestionsAnswers{}
}

// NewCommeLineParameters is parsing a list of strings to build a set of parameters
// for the AskQuestion function.
func Parse(args ...string) (InterrogationParameters, error) {
	p := InterrogationParameters{
		interactive: false,
		wait:        2 * time.Second,
		mode:        random,
		in:          os.Stdin,
		out:         os.Stdout,
		subsections: "",
		limit:       1,
		qachan:      make(chan string),
		command:     make(chan string),
		publisher:   make(chan string),
	}
	for i, opt := range args {
		switch opt {
		case "-i":
			p.interactive = true
		case "-t":
			value, err := strconv.Atoi(args[i+1])
			if err != nil {
				return p, fmt.Errorf("The time you set (%s) is not an integer. Please set the time in milliseconds.", args[i+1])
			}
			p.wait = time.Duration(value) * time.Millisecond
		case "-m":
			// The other mode is the default so we have nothing to do.
			if args[i+1] == "linear" {
				p.mode = linear
			}
		case "-s":
			p.mode = summary
		case "-l":
			p.subsections = args[i+1]
		case "-r":
			p.reversed = true
		}
	}
	return p, nil
}

// GetCount returns the number of entries for the questions.
func (qa QuestionsAnswers) GetCount() int {
	size := 0
	if qa.questions != nil {
		size = len(qa.questions)
	}
	return size
}

// NewTopic creates a new topic. Understand a topic as a set of questions
// with a title.
func NewTopic() Topic {
	return Topic{
		list: make(map[string]QuestionsAnswers),
	}
}

// GetSection returns the current list of questions for a given topic id.
// If there is no associated questions and answers for this topic id, it
// returns a new structure.
func (topic *Topic) GetSubsection(id string) QuestionsAnswers {
	qa := topic.list[id]
	if qa.questions == nil {
		qa = NewQA()
		topic.list[id] = qa
	}
	return qa
}

// SetSubsection defines a subsection with a given id and associates
// to it a list of questions.
func (topic *Topic) SetSubsection(id string, qa QuestionsAnswers) {
	topic.list[id] = qa
}

// GetCount returns the number of subtopics.
func (topic Topic) GetSubsectionsCount() int {
	size := 0
	if topic.list != nil {
		size = len(topic.list)
	}
	return size
}

// GetSubTopics returns the list of subtopics that have been imported.
func (topic Topic) GetSubsectionsName() []string {
	subsections := []string{}
	if topic.GetSubsectionsCount() != 0 {
		subsections = make([]string, 0, len(topic.list))
		for id := range topic.list {
			subsections = append(subsections, id)
		}
	}
	return subsections
}

// ParseQuestions is reading the data source and transforms it to a topic
// structure.
func ParseTopic(r io.Reader, p TopicParsingParameters) Topic {
	// Reading the file line by line
	s := bufio.NewScanner(r)

	lines := make([]string, 50)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	topic := NewTopic()
	var subsectionId string
	qaSubsection := NewQA()
	for i := 0; i < len(lines); i++ {
		input := lines[i]
		// Ignore empty lines
		if len(input) > 0 {
			split := strings.Split(input, p.QaSep)
			switch len(split) {
			case 1:
				if strings.HasPrefix(input, p.TopicAnnounce) {
					subsectionId = strings.TrimPrefix(input, p.TopicAnnounce)
					qaSubsection = topic.GetSubsection(subsectionId)
				}
			default:
				// Question is in split[0] while answer in in split[1]. It may happen
				// the answer contains the separator so we have to join the different
				// elements.
				qaSubsection.AddEntry(split[0], strings.Join(split[1:], p.QaSep))
				topic.SetSubsection(subsectionId, qaSubsection)
			}
		}
	}
	return topic
}

// AddEntry adds a set of question/answer to the already existing set.
func (qa *QuestionsAnswers) AddEntry(q string, a string) {
	qa.questions = append(qa.questions, q)
	qa.answers = append(qa.answers, a)
}

// Concatenate adds the entries of the parameter to an existing QA set.
func (qa *QuestionsAnswers) Concatenate(qaToAdd ...QuestionsAnswers) {
	var count int
	for _, toAdd := range qaToAdd {
		count = toAdd.GetCount()
		if count > 0 {
			qa.questions = append(qa.questions, toAdd.questions...)
			qa.answers = append(qa.answers, toAdd.answers...)
		}
	}
}

// BuildQuestionsSet creates a set of questions based on a Topic. We use a
// variadic list of parameters to allow to supply as many as topic on which
// the user wants to be questionned. If she/he supplies nothing, we use the
// the whole topic.
func (topic Topic) BuildQuestionsSet(ids ...string) QuestionsAnswers {
	qa := NewQA()
	var qaForId QuestionsAnswers
	var subsections = ids
	if len(subsections) == 0 {
		fmt.Println("     *** You supplied no subsection, we take them all ***")
		subsections = topic.GetSubsectionsName()
	}
	for _, id := range subsections {
		qaForId = topic.GetSubsection(id)
		qa.Concatenate(qaForId)
	}

	return qa
}

// fanOutChannel reads from the readFrom channel and dispatch the elements
// to the writeTo channel. When reading from the readFrom channel breaks,
// we write to the stopper channel the name of the channel from which we
// cannot read anymore.
func fanOutChannel(wg *sync.WaitGroup, readFrom <-chan string, writeTo chan<- string) {
	defer wg.Done()

	for {
		select {
		case v, ok := <- readFrom:
			if !ok {
				return
			}
			if len(v) != 0 {
				writeTo <- v
			}
		}
	}
}

// 
func publishChanToWriter(wg *sync.WaitGroup, readFrom <-chan string, out io.Writer, qCount int, maxLoops int) {
	defer wg.Done()
	itemsRead := 0
	currentLoop := 0
	c := color.New(color.FgBlue).Add(color.Bold)

	fmt.Fprintf(out, "Nb of questions: %d\n", qCount)

	for {
		if itemsRead%(2*qCount) == 0 {
			currentLoop++
			if currentLoop > maxLoops {
				fmt.Fprintf(out, "Limit reached. Exiting. Number of loops set to: %d\n",maxLoops)
				return
			}
			fmt.Fprintf(out, c.Sprintf("Loop (%d/%d)\n", currentLoop, maxLoops))
		}
		select {
		case v, ok := <- readFrom:
			if !ok {
				return
			}
			itemsRead++
			switch {
			case itemsRead%2==1:
				fmt.Fprintf(out, v)
				// Questions asked. Must publish the answer now.
			case itemsRead%2==0:
				fmt.Fprintf(out, "     --> " +v+"\n")
				fmt.Fprintf(out, "---------------------------\n")
			}
		}
	}
}


// AskQuestions will question the user on the set of questions. The
// parameter object will supply data to refine the questioning.
func AskQuestions(qa QuestionsAnswers, p InterrogationParameters) {
	fullLoop, i, j := 0, 0, 0

	var wg sync.WaitGroup
	wg.Add(3)
	nbOfQuestions := qa.GetCount()

	go fanOutChannel(&wg, p.qachan, p.publisher)
	go publishChanToWriter(&wg, p.publisher, p.out, nbOfQuestions, p.limit)
  go fanOutChannel(&wg, p.command, p.publisher)

	var question, answer string
	s := bufio.NewScanner(p.in)
	for {
		if j%nbOfQuestions == 0 {
			fullLoop++
			if fullLoop > p.limit {
				// if the qa chan is closed, then we have to close the others.
				close(p.qachan)
				close(p.command)
				break
			}
		}
		if p.mode == random {
			i = int(rand.Int31n(int32(nbOfQuestions)))
		}
		question = qa.questions[i]
		answer = qa.answers[i]
		if p.IsReversedMode() {
			question = qa.answers[i]
			answer = qa.questions[i]
		}
		p.qachan <- fmt.Sprintf("%s", question)
		if !p.interactive {
			time.Sleep(p.wait)
		} else {
			if s.Scan() {
				p.command <- s.Text()
			}
		}
		p.qachan <- fmt.Sprintf("%s", answer)

		if p.mode == linear {
			i = (i + 1) % nbOfQuestions
		}
		j++
	}

	wg.Wait()
}
