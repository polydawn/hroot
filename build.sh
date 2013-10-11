# Where is this script located?
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Build the project
(
	cd $DIR ;
	export GOPATH="$PWD"/.gopath/ ;
	go build -race -o docket polydawn.net/docket/main
)
