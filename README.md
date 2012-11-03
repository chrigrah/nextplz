nextplz
=======

nextplz is a simple Go application for browsing and playing video files.

Observe: At this time the application looks for the executable 'vlc' on the system path and uses that for a media player.
Controls:
	arrow keys:
	ctrl+{hjkl}:
		Move selection

	ctrl+backspace:
		Move up one directory level

	Enter:
		Enter the currently selected directory

	ctrl+n:
		Move to the "next" directory

	ctrl+p:
		Move to the "previous" directory
	
	ctrl+b:
		Play the currently selected media file

	Escape:
		Clear the input line/close application
