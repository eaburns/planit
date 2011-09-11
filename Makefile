include $(GOROOT)/src/Make.inc

TARG=goplan

GOFILES=\
	lex.go\
	parse.go\
	ast.go\
	print.go\
	main.go\

include $(GOROOT)/src/Make.cmd
