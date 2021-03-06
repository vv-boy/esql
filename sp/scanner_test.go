package sp_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/chenyoufu/esql/sp"
)

// errstring converts an error to its string representation.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// Ensure the scanner can scan tokens correctly.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s   string
		tok sp.Token
		lit string
		pos sp.Pos
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, tok: sp.EOF},
		{s: `#`, tok: sp.ILLEGAL, lit: `#`},
		{s: ` `, tok: sp.WS, lit: " "},
		{s: "\t", tok: sp.WS, lit: "\t"},
		{s: "\n", tok: sp.WS, lit: "\n"},
		{s: "\r", tok: sp.WS, lit: "\n"},
		{s: "\r\n", tok: sp.WS, lit: "\n"},
		{s: "\rX", tok: sp.WS, lit: "\n"},
		{s: "\n\r", tok: sp.WS, lit: "\n\n"},
		{s: " \n\t \r\n\t", tok: sp.WS, lit: " \n\t \n\t"},
		{s: " foo", tok: sp.WS, lit: " "},

		// Numeric operators
		{s: `+`, tok: sp.ADD},
		{s: `-`, tok: sp.SUB},
		{s: `*`, tok: sp.MUL},
		{s: `/`, tok: sp.DIV},

		// Logical operators
		{s: `AND`, tok: sp.AND},
		{s: `and`, tok: sp.AND},
		{s: `OR`, tok: sp.OR},
		{s: `or`, tok: sp.OR},

		{s: `=`, tok: sp.EQ},
		{s: `<>`, tok: sp.NEQ},
		{s: `! `, tok: sp.ILLEGAL, lit: "!"},
		{s: `<`, tok: sp.LT},
		{s: `<=`, tok: sp.LTE},
		{s: `>`, tok: sp.GT},
		{s: `>=`, tok: sp.GTE},

		// Misc tokens
		{s: `(`, tok: sp.LPAREN},
		{s: `)`, tok: sp.RPAREN},
		{s: `,`, tok: sp.COMMA},
		{s: `.`, tok: sp.DOT},
		{s: `=~`, tok: sp.EQREGEX},
		{s: `!~`, tok: sp.NEQREGEX},

		// Identifiers
		{s: `foo`, tok: sp.IDENT, lit: `foo`},
		{s: `_foo`, tok: sp.IDENT, lit: `_foo`},
		{s: `@foo`, tok: sp.IDENT, lit: `@foo`},
		{s: `Zx12_3U_-`, tok: sp.IDENT, lit: `Zx12_3U_`},
		{s: `test"`, tok: sp.BADSTRING, lit: "", pos: sp.Pos{Line: 0, Char: 3}},
		{s: `"test`, tok: sp.BADSTRING, lit: `test`},

		{s: `true`, tok: sp.TRUE},
		{s: `false`, tok: sp.FALSE},

		// Strings
		{s: `"foo"`, tok: sp.STRING, lit: `foo`},
		{s: `"foo\\bar"`, tok: sp.STRING, lit: `foo\bar`},
		{s: `"foo\bar"`, tok: sp.BADESCAPE, lit: `\b`, pos: sp.Pos{Line: 0, Char: 5}},
		{s: `"foo\"bar\""`, tok: sp.STRING, lit: `foo"bar"`},
		{s: `'testing 123!'`, tok: sp.STRING, lit: `testing 123!`},
		{s: `'foo\nbar'`, tok: sp.STRING, lit: "foo\nbar"},
		{s: `'foo\\bar'`, tok: sp.STRING, lit: "foo\\bar"},
		{s: `'test`, tok: sp.BADSTRING, lit: `test`},
		{s: "'test\nfoo", tok: sp.BADSTRING, lit: `test`},
		{s: `'test\g'`, tok: sp.BADESCAPE, lit: `\g`, pos: sp.Pos{Line: 0, Char: 6}},

		// Numbers
		{s: `100`, tok: sp.INTEGER, lit: `100`},
		{s: `10.3`, tok: sp.NUMBER, lit: `10.3`},
		// Keywords
		{s: `AS`, tok: sp.AS},
		{s: `ASC`, tok: sp.ASC},
		{s: `BY`, tok: sp.BY},
		{s: `DESC`, tok: sp.DESC},
		{s: `FROM`, tok: sp.FROM},
		{s: `GROUP`, tok: sp.GROUP},
		{s: `HAVING`, tok: sp.HAVING},
		{s: `LIMIT`, tok: sp.LIMIT},
		{s: `ORDER`, tok: sp.ORDER},
		{s: `SELECT`, tok: sp.SELECT},
		{s: `WHERE`, tok: sp.WHERE},
		{s: `seLECT`, tok: sp.SELECT}, // case insensitive
	}

	for i, tt := range tests {
		s := sp.NewScanner(strings.NewReader(tt.s))
		tok, pos, lit := s.Scan()
		if tt.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tt.s, tt.tok, tok, lit)
		} else if tt.pos.Line != pos.Line || tt.pos.Char != pos.Char {
			t.Errorf("%d. %q pos mismatch: exp=%#v got=%#v", i, tt.s, tt.pos, pos)
		} else if tt.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s, tt.lit, lit)
		}
	}
}

// Ensure the scanner can scan a series of tokens correctly.
func TestScanner_Scan_Multi(t *testing.T) {
	type result struct {
		tok sp.Token
		pos sp.Pos
		lit string
	}
	exp := []result{
		{tok: sp.SELECT, pos: sp.Pos{Line: 0, Char: 0}, lit: ""},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 6}, lit: " "},
		{tok: sp.IDENT, pos: sp.Pos{Line: 0, Char: 7}, lit: "value"},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 12}, lit: " "},
		{tok: sp.FROM, pos: sp.Pos{Line: 0, Char: 13}, lit: ""},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 17}, lit: " "},
		{tok: sp.IDENT, pos: sp.Pos{Line: 0, Char: 18}, lit: "myseries"},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 26}, lit: " "},
		{tok: sp.WHERE, pos: sp.Pos{Line: 0, Char: 27}, lit: ""},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 32}, lit: " "},
		{tok: sp.IDENT, pos: sp.Pos{Line: 0, Char: 33}, lit: "a"},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 34}, lit: " "},
		{tok: sp.EQ, pos: sp.Pos{Line: 0, Char: 35}, lit: ""},
		{tok: sp.WS, pos: sp.Pos{Line: 0, Char: 36}, lit: " "},
		{tok: sp.STRING, pos: sp.Pos{Line: 0, Char: 36}, lit: "b"},
		{tok: sp.EOF, pos: sp.Pos{Line: 0, Char: 40}, lit: ""},
	}

	// Create a scanner.
	v := `SELECT value from myseries WHERE a = 'b'`
	s := sp.NewScanner(strings.NewReader(v))

	// Continually scan until we reach the end.
	var act []result
	for {
		tok, pos, lit := s.Scan()
		act = append(act, result{tok, pos, lit})
		if tok == sp.EOF {
			break
		}
	}

	// Verify the token counts match.
	if len(exp) != len(act) {
		t.Fatalf("token count mismatch: exp=%d, got=%d", len(exp), len(act))
	}

	// Verify each token matches.
	for i := range exp {
		if !reflect.DeepEqual(exp[i], act[i]) {
			t.Fatalf("%d. token mismatch:\n\nexp=%#v\n\ngot=%#v", i, exp[i], act[i])
		}
	}
}

// Ensure the library can correctly scan strings.
func TestScanString(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err string
	}{
		{in: `""`, out: ``},
		{in: `"foo bar"`, out: `foo bar`},
		{in: `'foo bar'`, out: `foo bar`},
		{in: `"foo\nbar"`, out: "foo\nbar"},
		{in: `"foo\\bar"`, out: `foo\bar`},
		{in: `"foo\"bar"`, out: `foo"bar`},
		{in: `'foo\'bar'`, out: `foo'bar`},

		{in: `"foo` + "\n", out: `foo`, err: "bad string"}, // newline in string
		{in: `"foo`, out: `foo`, err: "bad string"},        // unclosed quotes
		{in: `"foo\xbar"`, out: `\x`, err: "bad escape"},   // invalid escape
	}

	for i, tt := range tests {
		out, err := sp.ScanString(strings.NewReader(tt.in))
		if tt.err != errstring(err) {
			t.Errorf("%d. %s: error: exp=%s, got=%s", i, tt.in, tt.err, err)
		} else if tt.out != out {
			t.Errorf("%d. %s: out: exp=%s, got=%s", i, tt.in, tt.out, out)
		}
	}
}

// Test scanning number
func TestScanNumber(t *testing.T) {
	var tests = []struct {
		s   string
		tok sp.Token
		lit string
	}{
		// Numbers
		{s: `0`, tok: sp.INTEGER, lit: `0`},
		{s: `000.0000`, tok: sp.NUMBER, lit: `000.0000`},
		{s: `100`, tok: sp.INTEGER, lit: `100`},
		{s: `10.3`, tok: sp.NUMBER, lit: `10.3`},
	}

	for i, tt := range tests {
		s := sp.NewScanner(strings.NewReader(tt.s))
		tok, _, lit := s.Scan()
		if tt.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tt.s, tt.tok, tok, lit)
		} else if tt.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s, tt.lit, lit)
		}
	}
}

// Test scanning regex
func TestScanRegex(t *testing.T) {
	var tests = []struct {
		in  string
		tok sp.Token
		lit string
		err string
	}{
		{in: `/^payments\./`, tok: sp.REGEX, lit: `^payments\.`},
		{in: `/foo\/bar/`, tok: sp.REGEX, lit: `foo/bar`},
		{in: `/foo\\/bar/`, tok: sp.REGEX, lit: `foo\/bar`},
		{in: `/foo\\bar/`, tok: sp.REGEX, lit: `foo\\bar`},
		{in: `/http\:\/\/www\.example\.com/`, tok: sp.REGEX, lit: `http\://www\.example\.com`},
	}

	for i, tt := range tests {
		s := sp.NewScanner(strings.NewReader(tt.in))
		tok, _, lit := s.ScanRegex()
		if tok != tt.tok {
			t.Errorf("%d. %s: error:\n\texp=%s\n\tgot=%s\n", i, tt.in, tt.tok.String(), tok.String())
		}
		if lit != tt.lit {
			t.Errorf("%d. %s: error:\n\texp=%s\n\tgot=%s\n", i, tt.in, tt.lit, lit)
		}
	}
}
