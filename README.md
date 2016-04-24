go is a really fucking stupid design by a bunch retarded geniuses at Google. 

In order to make it, we have to put everything in a "src" folder, then set the GOPATH environment variable to the parent folder of "src". e.g. in Ubuntu, type: mkdir src && export GOPATH=\`pwd\`

Then go to the project folder and type make. All the executable files will be created in the "bin" folder