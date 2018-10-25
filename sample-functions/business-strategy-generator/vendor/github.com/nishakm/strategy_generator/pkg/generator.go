package pkg

import (
	"fmt"
	"math/rand"
	"time"
)

// Statement is a struct and something to encapsulate the
// string and the type of word that gets filled in
// note that we are using Capital letters for the variables
// this is so we can access them directly
type Statement struct {
	Sentence string // the sentence to substitute a word in
	WordType int    // 0 for noun and 1 for adjective (more word organization needed! Pull Requests accepted :)
}

// nouns is an array assignment with no specific limits
// that means other words can be added here
var nouns = []string{
	"efficiency",
	"competitive advantage",
	"ecosystems",
	"synergy",
	"learning organization",
	"network",
	"social media",
	"revolution",
	"big data",
	"security",
	"internet of things",
	"digital business",
	"data leaders",
	"big data",
	"insight from data",
	"platform",
	"culture"}

var adjectives = []string{
	"digital first",
	"agile",
	"open",
	"innovative",
	"networked",
	"collaborative",
	"cloud based",
	"growth focused",
	"secure",
	"customer focused",
	"disruptive",
	"platform based",
	"sustainable",
	"value added"}

// Each statement has a %s which is for string substitution
// in Go there is no tuple substitution like python so the
// whole statement needs to be divided in this way
var statements = []Statement{
	Statement{"Our strategy is %s. ", 1},
	Statement{"We will lead a %s ", 1},
	Statement{"effort of the market through our use of %s ", 0},
	Statement{"and %s ", 0},
	Statement{"to build a %s. ", 0},
	Statement{"By being both %s ", 1},
	Statement{"and %s, ", 1},
	Statement{"our %s ", 1},
	Statement{"approach will drive %s throughout the organization. ", 0},
	Statement{"Synergies between our %s ", 0},
	Statement{"and %s ", 0},
	Statement{"will enable us to capture the upside by becoming %s ", 1},
	Statement{"in a %s world. ", 0},
	Statement{"These transformations combined with %s ", 0},
	Statement{"due to our %s ", 0},
	Statement{"will create a %s ", 0},
	Statement{"through %s ", 0},
	Statement{"and %s.", 0}}

// Generate produces a random business statedgy statement.
// Note: any of the functions called from the packages imported
// has to start with a Capital letter this is because of the rule
// that only functions starting with a capital letter are 'exported'
// so you can call them here from another package.
// It's not really the same as public and private but more like direct
// and indirect reference or use
// https://www.ardanlabs.com/blog/2014/03/exportedunexported-identifiers-in-go.html
func Generate() string {

	// implicit variable assignment
	// not available outside the function
	statement := ""

	// explicit unassigned variable
	var word string

	// rand uses a global package called Source
	// https://golang.org/pkg/math/rand/#Source
	// The source has to be initialized by some seed
	// the seed needs to change every time a random number needs to
	// be generated or else the same number will be generated
	for i := 0; i < len(statements); i++ {
		rand.Seed(time.Now().UnixNano())
		if statements[i].WordType == 0 {
			word = nouns[rand.Intn(len(nouns))]
		} else { //like C else has to be right after the closing }
			word = adjectives[rand.Intn(len(adjectives))]
		}
		statement = statement + fmt.Sprintf(statements[i].Sentence, word)
	}

	return statement
}
