package main

import (
	"fmt"
	"lib"
	"os"

	"github.com/fatih/color"
)

func main() {
	// Recuperation du parametre vers le fichier
	if len(os.Args) < 2 {
		c := color.New(color.FgRed).Add(color.Underline)
		c.Printf("Please supply a path to a CSV file that contains the topics.\n")

		c = color.New(color.FgWhite).Add(color.Bold)
		c.Printf(`Syntax:
	%s <csvFile> [-i]
where:
	* -i : stands for interactive. If set, you will have to press Return to get the
          answer. This allows you to be in a learning way or enforcing your knowledge.
			 If this flag is not set, you will not have to press the Return key and you
			 simply have to wait for a given time. See -t for details about time.
	* -t : the time to wait between 2 questions. Default is 2 seconds. The time you set is
	       in milliseconds.
	* -s : ask to show the different topics of  the file, no more. Execution stops after this.
	       Sections are supposed to start with ###.
	* -l : ask to be questionned only on the topics that are listed here. The topics must be separated with a comma.
	* -r : reverts the questioning. This is like a Jeopardy in fact. The right column becomes the questions while the right column becomes the answer.
`, os.Args[0])
		os.Exit(1)
	}

	// Creer un objet fichier et tester si on peut le lire
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Open of the source file failed: %v\n", err)
		os.Exit(1)
	}

	p, err := lib.Parse(os.Args[2:]...)
	if err != nil {
		fmt.Errorf("Parse of the command line failed: %v\n", err)
		os.Exit(1)
	}

	tpp := lib.TopicParsingParameters{
		TopicAnnounce: "### ",
		QaSep:         ";",
	}
	topic := lib.ParseTopic(file, tpp)
	file.Close()

	out := p.GetOutputStream()
	if p.IsSummaryMode() {
		list := topic.GetSubsectionsName()
		if len(list) == 0 {
			fmt.Fprintf(out, "No topic found in this file")
			return
		}
		fmt.Fprintln(out, "List of topics:")
		fmt.Fprintln(out, "===============")
		for i := 0; i < len(list); i++ {
			fmt.Fprintf(out, "  * %s\n", list[i])
		}
		return
	}

	qa := topic.BuildQuestionsSet(p.GetListOfSubsections()[:]...)

	lib.AskQuestions(qa, p)

}
