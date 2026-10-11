# Regenerate with: vhs assets/demo.tape (from the repository root)

Output assets/demo.gif

Set Shell "bash"
Set FontSize 22
Set Width 1440
Set Height 760
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type `export D=/tmp/my-tool && rm -rf "$D" && mkdir -p "$D" && go build -o "$D/mdschema" ./cmd/mdschema && cp assets/demo/.mdschema.yml assets/demo/README.md "$D" && cd "$D" && export PATH="$D:$PATH" PS1='\[\e[1;35m\]$ \[\e[0m\]' && clear`
Enter
Sleep 3s
Show

Type "cat .mdschema.yml"
Sleep 400ms
Enter
Sleep 3s

Type "mdschema check README.md"
Sleep 400ms
Enter
Sleep 4s

Type "clear"
Enter
Type "mdschema generate -o NEW.md && mdschema check NEW.md"
Sleep 400ms
Enter
Sleep 2s
Type "cat NEW.md"
Sleep 400ms
Enter
Sleep 4s
