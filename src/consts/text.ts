import { FOLDED_CHARACTER_MAP_LATIN } from './text/latin.ts';
import { FOLDED_CHARACTER_MAP_GREEK_CYRILLIC } from './text/greek-cyrillic.ts';
import { FOLDED_CHARACTER_MAP_EXTENDED_ALPHABETS } from './text/extended-alphabets.ts';
import { FOLDED_CHARACTER_MAP_BMP_COMPATIBILITY } from './text/bmp-compatibility.ts';
import { FOLDED_CHARACTER_MAP_SUPPLEMENTARY } from './text/supplementary.ts';
// Unicode case folding data (CaseFolding-17.0.0.txt) from Unicode Consortium: https://www.unicode.org/Public/UNIDATA/CaseFolding.txt
// The following bash scripts is about generating the folded character map from the Unicode data
/*
#!/bin/bash

curl -s 'https://www.unicode.org/Public/UNIDATA/CaseFolding.txt' | egrep -o '^[^#].*' | awk -F ';' '
function ltrim(s) {
    sub(/^[ \t]+/, "", s);
    return s;
}
function rtrim(s) {
    sub(/[ \t]+$/, "", s);
    return s;
}
function trim(s) {
    return rtrim(ltrim(s));
}
BEGIN {
    SINGLE_QUOTE_CHAR = 39;
    print "export const FOLDED_CHARACTER_MAP: Record<string, string> = {";
}
{
    if (trim($2) == "C" || trim($2) == "F") {
        replaced_str = trim($3);
        replaced_str = gensub(/ /, "\\\\u", "g", replaced_str);
        printf "    %c\\u%s%c: %c\\u%s%c,\n", SINGLE_QUOTE_CHAR, trim($1), SINGLE_QUOTE_CHAR, SINGLE_QUOTE_CHAR, replaced_str, SINGLE_QUOTE_CHAR;
    }
}
END {
    print "};";
}'
*/
export const FOLDED_CHARACTER_MAP: Record<string, string> = {
    ...FOLDED_CHARACTER_MAP_LATIN,
    ...FOLDED_CHARACTER_MAP_GREEK_CYRILLIC,
    ...FOLDED_CHARACTER_MAP_EXTENDED_ALPHABETS,
    ...FOLDED_CHARACTER_MAP_BMP_COMPATIBILITY,
    ...FOLDED_CHARACTER_MAP_SUPPLEMENTARY,
};
