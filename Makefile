include $(GOROOT)/src/Make.inc

TARG=goplan

GOFILES=\
	lex.go\
	parse.go\
	print.go\
	domain.go\
	main.go\

include $(GOROOT)/src/Make.cmd
