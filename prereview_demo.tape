# VHS Demo for PreReview
# Run with: vhs demo.tape

Output prereview_demo.gif

Set Theme "Dracula"
Set FontSize 14
Set Width 900
Set Height 500
Set TypingSpeed 50ms
Set Padding 20

# Start fresh
Type "# Let's review some staged changes with prereview"
Enter
Sleep 500ms

Type "git status"
Enter
Sleep 1s

Type "prereview --model claude"
Enter
Sleep 8s

# Show the review happening
Sleep 3s

# End with a nice pause
Sleep 2s
