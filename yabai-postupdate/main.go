package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strings"
)

func main() {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	bYabai, err := exec.Command("which", "yabai").Output()
	if err != nil {
		log.Fatal(err)
	}
	yabai, _ := strings.CutSuffix(string(bYabai), "\n")

	bSha, err := exec.Command("shasum", "-a", "256", yabai).Output()
	if err != nil {
		log.Fatal(err)
	}
	sha, _ := strings.CutSuffix(string(bSha), "\n")

	f := fmt.Sprintf("%s ALL=(root) NOPASSWD: sha256:%s --load-sa", u.Username, sha)

	// err = os.WriteFile("/private/etc/sudoers.d/yabai", []byte(f), 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	cmd := exec.Command("sudo", "tee", "/private/etc/sudoers.d/yabai")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdin = bytes.NewBufferString(f)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("yabai", "--restart-service").Run()
	if err != nil {
		log.Fatal(err)
	}
}
