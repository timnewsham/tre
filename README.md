# tim re

some regexp stuff

## Syntax

```
re :=
    .               # matches any character, including newline
    [ cclass ]      # matches characters in the character class
    [^ cclass ]     # matches characters not in the character class
    ch              # matches ch if it is not a metacharacter
    \ ch            # matches ch directly if it is punctuation, newline for \n, carriage return for \r.
    ( re )          # matches re
    (? re )         # matches re and greedily captures the matching string.
    re ?            # matches zero or one re
    re *            # matches zero or more re
    re +            # matches one or more re
    re re           # matches first re followed by second re
    re | re         # matches first re or second re

cclass :=
    ch              # matches character if it is not a metacharacter
    \ ch            # matches ch directly if it is puncutation, newline for \n, carriage return for \r.
    ch-ch           # matches any character from first ch to second ch, inclusively. second ch cannot be less than first ch.
    cclass cclass   # matches character in first or second cclass.

bounded re :=
    terminal re terminal    # terminal becomes a metacharacter in re, and the same terminal must end the bounded re.
```
