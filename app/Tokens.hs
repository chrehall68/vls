module Tokens where

data Token
  = -- main things
    IDENTIFIER String
  | LITERAL String
  | TYPE String
  | -- keywords
    MODULE
  | ENDMODULE
  | BEGIN
  | END
  | CASE
  | ENDCASE
  | GENERATE
  | ENDGENERATE
  | FOR
  | IF
  | ELSE
  | ASSIGN
  | INITIAL
  | TIME
  | DEFAULT
  | -- comparisons/assignments
    COMPARATOR String
  | LOGICAL_OPERATOR String
  | OPERATOR String
  | -- symbols
    LPAREN
  | RPAREN
  | LBRACKET
  | RBRACKET
  | LCURL
  | RCURL
  | COLON
  | COMMA
  | DOT
  | SEMICOLON
  | BACKTICK
  | QUESTION
  | AT
  | EQUAL
  | TILDA
  | POUND
  | UNKNOWN String
  deriving (Show, Eq)
