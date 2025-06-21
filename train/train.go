package train

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/mb-14/gomarkov"
)

const (
	hnBaseURL        = "https://hacker-news.firebaseio.com/v0/"
	hnTopStoriesPath = "topstories.json"
	hnStoryItemPath  = "item/"
)

type hnStory struct {
	Title string `json:"title"`
}

// func main() {
// 	train := flag.Bool("train", false, "Train the markov chain")
// 	flag.Parse()
// 	if *train {
// 		chain, err := buildModel()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		saveModel(chain)
// 	} else {
// 		chain, err := loadModel()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		generateHNStory(chain)
// 	}
// }

func BuildModel(input string) (*gomarkov.Chain, error) {
	chain := gomarkov.NewChain(1)
	//i should probably split out punctionation, todo

	err := AddTextToModel(chain, input)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

// AddTextToModel adds additional text to an existing markov chain model
func AddTextToModel(chain *gomarkov.Chain, input string) error {
	simple := true
	if simple {
		paragraphs := strings.Split(input, "\n\n")
		for _, paragraph := range paragraphs {
			worksWithPunctuation := strings.Split(paragraph, " ")
			chain.Add(worksWithPunctuation)
		}
	} else {
		//now extract punctuation and add to the chain separately
		//also consider quotes and other punctuation that can start a sentence
		//also consider whitespace as it's own token

		worksWithPunctuation := strings.Split(input, " ")
		punctuation := []string{".", "!", "?", ","}
		var tokens []string
		for _, word := range worksWithPunctuation {
			for _, punct := range punctuation {
				if strings.HasSuffix(word, punct) {
					tokens = append(tokens, strings.TrimSuffix(word, punct), punct)
				} else {
					tokens = append(tokens, word)
				}
				tokens = append(tokens, " ")
			}
		}
		chain.Add(tokens)
	}
	return nil
}

func LoadModel(data []byte) (*gomarkov.Chain, error) {
	var chain gomarkov.Chain
	err := json.Unmarshal(data, &chain)
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

func SerializeModel(chain *gomarkov.Chain) ([]byte, error) {
	return json.Marshal(chain)
}

func GenerateStory(prngSeed int64, chain *gomarkov.Chain) (string, *rand.Rand, error) {
	prng := rand.New(rand.NewSource(prngSeed))
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.GenerateDeterministic(tokens[(len(tokens)-1):], prng)
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " "), prng, nil
}

func GenerateStoryFromPrng(prng *rand.Rand, chain *gomarkov.Chain) (string, error) {
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.GenerateDeterministic(tokens[(len(tokens)-1):], prng)
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " "), nil
}

func GenerateStoryBasic(chain *gomarkov.Chain) (string, error) {
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.Generate(tokens[(len(tokens) - 1):])
		fmt.Println(next)
		// time.Sleep(100 * time.Millisecond)
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " "), nil
}

// SeedablePRNG implements a simple pseudo-random number generator with seed support
type SeedablePRNG struct {
	seed  int64
	state int64
}

// NewSeedablePRNG creates a new PRNG with the given seed
func NewSeedablePRNG(seed int64) *SeedablePRNG {
	return &SeedablePRNG{
		seed:  seed,
		state: seed,
	}
}

// Intn returns a random number in the half-open interval [0,n)
func (p *SeedablePRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	// Simple linear congruential generator
	p.state = (p.state*1103515245 + 12345) & 0x7fffffff
	return int(p.state % int64(n))
}

// Reset resets the PRNG to its initial state
func (p *SeedablePRNG) Reset() {
	p.state = p.seed
}

func (p *SeedablePRNG) GetSeed() int64 {
	return p.state
}
