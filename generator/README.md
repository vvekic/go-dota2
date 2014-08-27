go-dota2's generator.go is based upon [go-steam](https://github.com/Philipp15b/go-steam)'s [generator.go](https://github.com/Philipp15b/go-steam/blob/master/generator/generator.go), and is subject to [go-steam's BSD license](https://github.com/Philipp15b/go-steam/blob/master/LICENSE.txt).

#get steamkit submodule
    git submodule update --init --recursive

#windows
    go run generator.go

#linux
    go run generator.go clean proto
