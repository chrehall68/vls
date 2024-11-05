{
module Scanner  where

import Tokens (Token(LPAREN, RPAREN, LBRACKET, RBRACKET, LCURL, RCURL, COLON, COMMA, DOT, SEMICOLON, COMPARATOR, LOGICAL_OPERATOR, OPERATOR, MODULE, ENDMODULE, BEGIN, END, CASE, ENDCASE, GENERATE, ENDGENERATE, FOR, IF, ELSE, ASSIGN, IDENTIFIER, LITERAL, TYPE, UNKNOWN, BACKTICK, ASSIGN, INITIAL, QUESTION, TIME, AT, DEFAULT, EQUAL, TILDA, POUND))
}

%wrapper "basic"

-- Add your token regular expressions and associated actions below.
tokens :-
  -- whitespace doesn't matter
  $white+  ;
  -- neither do comments
  \/\/.*  ;
  -- keywords
  module  { \s -> MODULE }
  endmodule {\s -> ENDMODULE}
  begin {\s -> BEGIN}
  end {\s -> END}
  case {\s -> CASE}
  endcase {\s->ENDCASE}
  generate {\s -> GENERATE}
  endgenerate {\s -> ENDGENERATE }
  for {\s->FOR}
  if {\s->IF}
  else {\s->ELSE}
  assign {\s->ASSIGN}
  initial {\s->INITIAL}
  (negedge)|(posedge) {\s->TIME}
  default {\s->DEFAULT}
  -- comparisons/assignments
  (\=\=)|(\!\=)|(\<\=)|(>\=)|\>|\<|(\=\=\=)|(\!\-\=) {\s->COMPARATOR s}
  (\&\&)|(\|\|) {\s->LOGICAL_OPERATOR s}
  [\+\-\*\/\|&] {\s->OPERATOR s}
  -- symbols
  \( {\s->LPAREN}
  \) {\s->RPAREN}
  \[ {\s->LBRACKET}
  \] {\s->RBRACKET}
  \{ {\s->LCURL}
  \} {\s->RCURL}
  : {\s->COLON}
  \, {\s->COMMA}
  \. {\s->DOT}
  \; {\s->SEMICOLON}
  ` {\s->BACKTICK}
  \? {\s->QUESTION}
  @ {\s->AT}
  \= {\s->EQUAL}
  \~ {\s -> TILDA }
  \# {\s -> POUND}
  -- variable-related
  (reg)|(wire)|(genvar)|(parameter)|(input)|(output)|(defparam) {\s->TYPE s}
  [A-Za-z][a-zA-Z0-9_]* {\s -> IDENTIFIER s}
  ([0-9]+)|([0-9]*\'[hbd][0-9xzXZA-Fa-f]+)|(\"[^\s]*\") {\s->LITERAL s}
  `include.+ {\s -> UNKNOWN s}
  -- other...
  \$.* ;