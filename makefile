.PHONY: default clean

default:
	go build -race -ldflags="-w -s"
	./Palette
	rm Palette.exe