# Demo tape for linkedin. Rendered with ascii-gif (github.com/tamnd/ascii-gif),
# which supplies the window chrome and theme; this file is just the action.
#
#   ascii-gif render docs/demo/linkedin.tape -o docs/static/demo.gif
#
# linkedin must be on PATH inside the recording shell.

Hide
Type "export PS1='$ ' PATH=/Users/apple/bin:$PATH"
Enter
Type "clear"
Enter
Show

Sleep 500ms
Type "linkedin jobs 'golang backend' -n 5 --fields job_id,title,company,location,posted"
Sleep 800ms
Enter
Sleep 3s

Type "linkedin profile williamhgates --fields name,headline,followers,location"
Sleep 800ms
Enter
Sleep 2.5s

Type "linkedin jobs 'golang backend' -n 4 -o jsonl | jq -r .title"
Sleep 800ms
Enter
Sleep 3s
