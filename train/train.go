package train

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/mb-14/gomarkov"
)

type MarkovChain struct {
	chain *gomarkov.Chain
}

func BuildModel(input string) (MarkovChain, error) {
	chain := gomarkov.NewChain(1)
	//i should probably split out punctionation, todo
	chainOut := MarkovChain{chain: chain}
	err := AddTextToModel(chainOut, input)
	if err != nil {
		return MarkovChain{}, err
	}
	return chainOut, nil
}

// AddTextToModel adds additional text to an existing markov chain model
func AddTextToModel(chain MarkovChain, input string) error {
	terminatingPunctuation := []string{".", "!", "?"}
	// now loop over fields, grouping by sentence, meaning gorup until a period is found.
	words := strings.Fields(input)
	lastIndex := 0
	for i := 0; i < len(words); i++ {
		word := words[i]
		lastChar := word[len(word)-1:]
		if slices.Contains(terminatingPunctuation, lastChar) {
			chain.chain.Add(words[lastIndex : i+1])
			fmt.Println(strings.Join(words[lastIndex:i+1], " "))
			lastIndex = i + 1
		}
	}
	if lastIndex < len(words) {
		chain.chain.Add(words[lastIndex:])
	}

	return nil
}

func LoadModel(data []byte) (MarkovChain, error) {
	var chain gomarkov.Chain
	err := json.Unmarshal(data, &chain)
	if err != nil {
		return MarkovChain{}, err
	}
	return MarkovChain{chain: &chain}, nil
}

func SerializeModel(chain MarkovChain) ([]byte, error) {
	return json.Marshal(chain.chain)
}

func GenerateStory(prngSeed int64, chain MarkovChain) (string, *rand.Rand, error) {
	prng := rand.New(rand.NewSource(prngSeed))
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.chain.GenerateDeterministic(tokens[(len(tokens)-1):], prng)
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " "), prng, nil
}

func GenerateStoryFromPrng(prng *rand.Rand, chain MarkovChain) (string, error) {
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.chain.GenerateDeterministic(tokens[(len(tokens)-1):], prng)
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " "), nil
}

func GenerateStoryBasic(chain MarkovChain) (string, error) {
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.chain.Generate(tokens[(len(tokens) - 1):])
		fmt.Println(next)
		// time.Sleep(100 * time.Millisecond)
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " "), nil
}
