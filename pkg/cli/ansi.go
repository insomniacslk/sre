package cli

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/johnmccabe/img2ansi"
)

func bold(s string) string {
	return color.New(color.Bold).SprintFunc()(s)
}

func italic(s string) string {
	return color.New(color.Italic).SprintFunc()(s)
}

func toAnsiURL(text, url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

func toTerminalImage(img []byte) (string, error) {
	if os.Getenv("LC_TERMINAL") == "iTerm2" {
		// use iTerm2's image support, https://iterm2.com/documentation-images.html
		return fmt.Sprintf("\033]1337;File=inline=1:%s\a\n", base64.StdEncoding.EncodeToString(img)), nil
	} else {
		return toAnsiImage(img)
	}
}

func toAnsiImage(img []byte) (string, error) {
	i, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return "", fmt.Errorf("failed to parse image: %w", err)
	}

	ansi, err := img2ansi.RenderTrueColor(i)
	if err != nil {
		log.Fatalf("Error converting image to ANSI: %v", err)
	}
	return ansi, nil
}
