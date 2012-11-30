nextplz
=======

nextplz is a simple Go application for browsing and playing video files. It uses termbox-go, a terminal UI library, to provide a clear and simple to use interface. By providing recursive video listings, instant searching/filtering of files and folders, and keybindings for common operations, nextplz aims to provide an interface that is quick and easy to use.

Observe: At this time the application looks for the executable 'vlc' on the system path and uses that for a media player. If you want to use another media player, please look for at the instructions under advanced usage.

Controls
========

	arrow keys up&down:
	ctrl+{yuio}:
		Move selection

	Page up:
	'~':
	'§':
		Move up one directory level

	Enter:
		Enter the currently selected directory

	ctrl+n:
		Move to the "next" directory

	ctrl+p:
		Move to the "previous" directory
	
	ctrl+b:
		Play the currently selected media file

	F3:
		Change directory (and/or drive on windows)

	F4:
		Recursively list media files in current folder

	Escape:
		Magic


Advanced usage
==============
For some reason VLC will not queue files while it has a video paused, so nextplz can toggle pause in VLC for you with the ctrl+space key combination. For this to work VLC must have been started from nextplz, or otherwise been configured so that it is listening for commands on TCP port 47246.