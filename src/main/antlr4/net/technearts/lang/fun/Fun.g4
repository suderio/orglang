grammar Fun;

@parser::members {
    private int operatorNesting = 0;

    public boolean isInsideOperator() {
        return operatorNesting > 0;
    }
}

file                : (assign SEMICOLON)*                                                                   #fileTable
                    ;

assign              : ID ASSIGN right=expression                                                            #assignExp
                    | expression                                                                            #expressionExp
                    ;
expression          : LPAREN expression RPAREN                                                              #parenthesisExp
                    | LCURBR { operatorNesting++; } op=expression RCURBR { operatorNesting--; }             #operatorExp
                    | left=expression DEREF right=expression                                                #derefExp
                    | <assoc=right> (PLUS|MINUS|NOT|INC|DEC) right=expression                               #unaryExp
                    | <assoc=right> left=expression EXP right=expression                                    #powerExp
                    | left=expression (ASTERISK|SLASH|PERCENT) right=expression                             #mulDivModExp
                    | left=expression (PLUS|MINUS) right=expression                                         #addSubExp
                    | left=expression DOLAR right=expression                                                #substExp
                    | left=expression (RSHIFT|LSHIFT) right=expression                                      #shiftExp
                    | left=expression (LT|LE|GE|GT) right=expression                                        #comparisonExp
                    | left=expression (EQ|NE) right=expression                                              #equalityExp
                    | left=expression AND_SHORT right=expression                                            #andShortExp
                    | left=expression AND right=expression                                                  #andExp
                    | left=expression XOR right=expression                                                  #xorExp
                    | left=expression OR_SHORT right=expression                                             #orShortExp
                    | left=expression OR right=expression                                                   #orExp
                    | <assoc=right> left=expression NULLTEST right=expression                               #nullTestExp
                    | left=expression (ASUM|ASUB|AMULT|ADIV|AMOD|ALSH|ARSH|AAND|AXOR|AOR) right=expression  #assignOpExp
                    | left=expression ID right=expression                                                   #biCallExp
                    | exp+=expression (SEPARATOR exp+=expression)+                                          #tableConcatSepExp
                    | LBRACK (exp+=expression)* RBRACK                                                      #tableConstructExp
                    | left=expression RANGE right=expression                                                #rangeExp
                    | ID ASSIGN_KEY right=expression                                                        #keyValueExp
                    | left=expression REDIRECT right=expression                                             #redirectWriteExp
                    | REDIRECT right=expression                                                             #redirectReadExp
                    | <assoc=right> left=expression TEST                                                    #testExp
                    | ID                                                                                    #idAtomExp
                    | ID right=expression                                                                   #callExp
                    | THIS right=expression                                                                 #thisExp
                    | SIMPLESTRING                                                                          #stringLiteral
                    | DOCSTRING                                                                             #docStringLiteral
                    | TRUE                                                                                  #trueLiteral
                    | FALSE                                                                                 #falseLiteral
                    | NULL                                                                                  #nullLiteral
                    | DECIMAL                                                                               #decimalLiteral
                    | INTEGER                                                                               #integerLiteral
                    | URL                                                                                   #urlLiteral
                    | {isInsideOperator()}? LEFT                                                            #leftAtomLiteral
                    | {isInsideOperator()}? RIGHT                                                           #rightAtomLiteral
                    ;

// Whitespace
NEWLINE             : '\r\n' | '\r' | '\n' ;
WS                  : [\r\n\t ]+ -> channel(HIDDEN) ;
COMMENT             : '#' [.]*? NEWLINE -> channel(HIDDEN)  ;
BLOCK_COMMENT       : NEWLINE? [ \t]* '###' .*? NEWLINE [ \t]* '###' -> channel(HIDDEN);

// Keywords
THIS                : 'this' ;
//IT                  : 'it'   ;
LEFT                : 'left'   ;
RIGHT               : 'right'   ;
TRUE                : 'true' ;
FALSE               : 'false';
NULL                : 'null' ;

// Literals
INTEGER            : [0-9]+;
DECIMAL            : '0'|[1-9][0-9]* '.' [0-9]+ ;
SIMPLESTRING       : '"' (~["\\] | '\\' .)* '"';
DOCSTRING          : '"""' .*? '"""';

// Operators

SEPARATOR          : ',' ;
ASSIGN             : ':' ;
ASSIGN_KEY         : '->' ;
LPAREN             : '(' ;
RPAREN             : ')' ;
LBRACK             : '[' ;
RBRACK             : ']' ;
LCURBR             : '{' ;
RCURBR             : '}' ;
SEMICOLON          : ';' ;

// Arithmetic
PLUS               : '+' ;
MINUS              : '-' ;
ASTERISK           : '*' ;
SLASH              : '/' ;
EXP                : '**';
PERCENT            : '%' ;

// Logical
NOT                : '~' ;
AND                : '&' ;
OR                 : '|' ;
AND_SHORT          : '&&';
OR_SHORT           : '||';
XOR                : '^' ;

// Comparison
EQ                 : '=' ;
LE                 : '<=';
LT                 : '<' ;
GE                 : '>=';
GT                 : '>' ;
NE                 : '<>' | '~=';

// OTHER
LSHIFT             : '<<';
RSHIFT             : '>>';
ASUM               : ':+';
ASUB               : ':-';
AMULT              : ':*';
ADIV               : ':/';
AMOD               : ':%';
ALSH               :':<<';
ARSH               :':>>';
AAND               : ':&';
AXOR               : ':^';
AOR                : ':|';
INC                : '++';
DEC                : '--';
RANGE              : '..';
DEREF              : '.' ;
REDIRECT           : '@' ;
DOLAR              : '$' ;
TEST               : '?' ;
NULLTEST           : '??';
ELVIS              : '?:';
MAP                :'-->';
FILTER             :'-|-';
REDUCE             :'<--';


// Identifiers
ID                 : [_]+[A-Za-z0-9_]* | [A-Za-z]+[A-Za-z0-9_]*;
URL                : (('http' | 'https' | 'ftp' | 'file') '://' (~[ \n\r\t])+);

// Should not be used
ANY                : . ;

// <=> <-