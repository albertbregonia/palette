.PHONY: default clean

default:
	go build -race -ldflags="-w -s"
	./Palette
clean: 
	rm Palette.exe