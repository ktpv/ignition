#!/usr/bin/env bash -el

notes=()

LOGFILE=/tmp/setup-ignition.log

if ! which brew >/dev/null 2>&1; then
    echo "You must have home brew installed" >&2
    echo "Visit https://brew.sh for more information" >&2
    exit 1
fi

if ! which go >/dev/null 2>&1; then
    echo "Installing Golang"
    brew install go >>$LOGFILE 2>&1
    notes+=('You must have a $GOPATH configured')
elif [[ ! -d "$GOPATH" ]]; then
    notes+=('You must have a $GOPATH configured')
fi

if ! which yarn >/dev/null 2>&1; then
    echo "Installing Yarn"
    brew install yarn >>$LOGFILE 2>&1
    notes+=("You need to add the Yarn Global Binary Path to your \$PATH\nexport PATH=\$PATH:\$(yarn global bin)")
fi

echo "Installing Dep"
go get -u github.com/golang/dep/cmd/dep

mkdir -p $(go env GOPATH)/src/github.com/pivotalservices
cd $(go env GOPATH)/src/github.com/pivotalservices

echo "Cloning Repo"
git clone git@github.com:pivotalservices/ignition >>$LOGFILE 2>&1
cd ignition
echo "Installing Go Dependencies"
dep ensure >>$LOGFILE 2>&1

pushd web >/dev/null 2>&1
    echo "Installing UI deps"
    yarn install >>$LOGFILE 2>&1
popd >/dev/null 2>&1

for note in ${notes[@]}; do
    echo -e "$note\n"
done

echo 'DONE!'
