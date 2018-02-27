package lib

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Lesson    = "### Lesson "
	Sentences = "### Sentences "
)

type QuestionsAnswers struct {
	questions []string
	answers   []string
}

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
type InterrogationParameters struct {
	interactive bool
	wait        time.Duration
}

func NewQA() QuestionsAnswers {
	return QuestionsAnswers{}
}

// NewCommeLineParameters is parsing a list of strings to build a set of parameters
// for the AskQuestion function.
func Parse(args ...string) (InterrogationParameters, error) {
	p := InterrogationParameters{
		interactive: false,
		wait:        2 * time.Second,
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

// NewTopic creates a new topic.
func NewTopic() Topic {
	return Topic{
		list: make(map[string]QuestionsAnswers),
	}
}

// GetSection returns the current list of questions for a given topic id.
// If there is no associated questions and answers for this topic id, it
// returns a new structure.
func (topic *Topic) GetSection(id string) QuestionsAnswers {
	qa := topic.list[id]
	if qa.questions == nil {
		qa = NewQA()
		topic.list[id] = qa
	}
	return qa
}

func (topic *Topic) SetSection(id string, qa QuestionsAnswers) {
	topic.list[id] = qa
}

// GetCount returns the number of subtopics.
func (topic Topic) GetCount() int {
	size := 0
	if topic.list != nil {
		size = len(topic.list)
	}
	return size
}

// GetSubTopics returns the list of subtopics that have been imported.
func (topic Topic) GetSubTopics() []string {
	subtopics := []string{}
	if topic.GetCount() != 0 {
		subtopics = make([]string, len(topic.list))
		for id := range topic.list {
			subtopics = append(subtopics, id)
		}
	}
	return subtopics
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
	var sectionId string
	qaSection := NewQA()
	for i := 0; i < len(lines); i++ {
		input := lines[i]
		// Ignore empty lines
		if len(input) > 0 {
			split := strings.Split(input, p.QaSep)
			switch len(split) {
			case 1:
				if strings.HasPrefix(input, p.TopicAnnounce) {
					sectionId = strings.TrimPrefix(input, p.TopicAnnounce)
					qaSection = topic.GetSection(sectionId)
				}
			case 2:
				// Question is in split[0] while answer in in split[1]
				qaSection.AddEntry(split[0], split[1])
				topic.SetSection(sectionId, qaSection)
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
	for _, id := range ids {
		fmt.Printf("L'element: %v\n", topic.GetSection(id))
		qaForId = topic.GetSection(id)
		qa.Concatenate(qaForId)
		fmt.Printf("Elements dans qaForId: %d\n", qaForId.GetCount())
		fmt.Printf("Count of elements: %d\n", qa.GetCount())
	}
	if len(ids) == 0 {
		// we must embed everything
		for _, id := range topic.GetSubTopics() {
			qaToAdd := topic.GetSection(id)
			qa.Concatenate(qaToAdd)
		}
	}

	return qa
}

// AskQuestions will question the user on the set of questions.
func AskQuestions(qa QuestionsAnswers, p InterrogationParameters) {
	// Interrogations en ordre aleatoire
	r := bufio.NewReader(os.Stdin)
	for {
		i := rand.Int31n(int32(qa.GetCount()))
		fmt.Printf("%s", qa.questions[i])
		if !p.interactive {
			time.Sleep(p.wait)
		} else {
			r.ReadLine()
		}
		fmt.Printf("     --> %s\n", qa.answers[i])
		fmt.Println("--------------------------")
	}
}
