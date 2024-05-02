# llm-bible-bench
A simple test for large language models and their recall on bible verses

## Text source

The original source of kjvdat.txt was lost, if found I will link it here

TODO: Switch to https://github.com/aruljohn/Bible-kjv

## How to use

* Install golang https://golang.org/
* Open a terminal, clone this project, set it to active directory
* Update packages with `go mod tidy`
* Run the program with `go run main.go`

## Configuration

* Configure the code in the function `getLLM()` to change the URL and authentication token of your LLM provider.
* Change the comments on the `BookNames` variable to change which books to run the test on. This is useful as the test may take some time to complete.

## Output

The program will output the overall accuracy, the percentage of verses that the language model correctly recalled.

It will also print CSV text that can be easily imported into excel and visualized with conditional cells.

See this example excel formatting of a test with llama3 on 2 Peter (11% accuracy)

![Example of visualized output from llama3 on 2 Peter](Example.png)

## Contribution

Contributions are welcome, please keep the code as simple as possible as this is meant to be a very quick and easy test to manage.