nextplz
=======

nextplz is a simple Go application for browsing and playing video files. It uses termbox-go, a terminal UI library, to provide a clear and simple to use interface. By providing recursive video listings, instant searching/filtering of files and folders, and keybindings for common operations, nextplz aims to provide an interface that is quick and easy to use.

Observe: At this time the application looks for the executable 'vlc' on the system path and uses that for a media player. If you want to use another media player, please look under Usage below.

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


Secret sauce
==============
For some reason VLC will not queue files while it has a video paused, so nextplz can toggle pause in VLC for you with the ctrl+space key combination. For this to work VLC must have been started from nextplz, or otherwise been configured so that it is listening for commands on TCP port 47246.

Usage
=====
  -args="": Arguments to be passed to the media player
  -cw=50: Column width for directory listing.

  -exe="": The name of the media player executable (must be on system path)
  -extensions=".avi,.mkv,.mpg,.wmv": Comma separated list of file extensions that should be considered video files.

  -filter-samples=true: If set to true, video files matching [.-]sample[.-] will be filtered out from recursive listings.
  -filter-subs=true: If set to true, rar files matching [.-]subs[.-] will be filtered out from recursive listings.
  -rar-folders=true: If set to true rar files will also be filtered by folder in recursive listings