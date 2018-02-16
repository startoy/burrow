## Marmot Hunt

If you are here you have probably been selected as a potential marmot. We have set a problem based on the existing Burrow code in the 'marmot' directory. In this package there is a problem we would like you to attempt in the guise of a broken unit test `TestMarmotAuthorisation`. The challenge is to make this unit test pass.

The goal in setting this problem is both to gain some insight about how you go about solving a problem and to provide a introduction to the code you would work on in platform team and and the problem space more generally. As such please include comment explaining how you arrived at your solution (or if you did not - an insightful explanation will still beat a solution with seemingly no insight) and what you have learnt about the system. We'd also welcome questions or comments on the codebase. If you have written any exploratory code or helper functions when investigating the system please include them as part of your submission - even if they are not used.

The goal is for you to demonstrate your competence and to provide a framework for a technical conversation and not for total perfection, so while we would like you to explain things you have learnt about the system - an exhaustive description is not necessary, and we do not want you to spend more than a few hours on the problem. Having said this we will use your submission to judge your suitability as a candidate for the role in question - as imperfect as that may be.

### Cloning the repo
You will need a working Go development environment to run the test see: [https://golang.org/doc/install](https://golang.org/doc/install).

The repo needs to be cloned into the proper namespace so:
```shell
mkdir -p $GOPATH/src/github.com/hyperledger/burrow/
git clone git@github.com:monax/monax.git $GOPATH/src/github.com/hyperledger/burrow/
git checkout marmot-hunt

```

Or to set origin correctly:
```
go get github.com/hyperledger/burrow
cd $GOPATH/src/github.com/hyperledger/burrow/
git remote add monax git@github.com:monax/monax.git
git fetch monax
git checkout marmot-hunt

```

### Running test

To run the test use:

```shell
go test ./marmot -run TestMarmotAuthorisation
```
### Submission

Clone the repo, start a new branch, add exploratory code and your solution, commit the code then use `git format-patch` to produce a plain text patch that you can then email to us.
