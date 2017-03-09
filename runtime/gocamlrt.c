#include <stdio.h>
#include <inttypes.h>
#include <stdlib.h>
#include <string.h>
#include "gocaml.h"

void print_int(gocaml_int const i)
{
    printf("%" PRId64, i);
}

void print_bool(gocaml_bool const i)
{
    printf("%s", i ? "true" : "false");
}

void print_float(gocaml_float const d)
{
    printf("%lg", d);
}

void print_str(gocaml_string const s)
{
    printf("%s", s.chars);
}

void println_int(gocaml_int const i)
{
    printf("%" PRId64 "\n", i);
}

void println_bool(gocaml_bool const i)
{
    printf("%s\n", i ? "true" : "false");
}

void println_float(gocaml_float const d)
{
    printf("%lg\n", d);
}

void println_str(gocaml_string const s)
{
    printf("%s\n", s.chars);
}

gocaml_int float_to_int(gocaml_float const f)
{
    return (gocaml_int) f;
}

gocaml_float float_of_int(gocaml_int const i)
{
    return (gocaml_float) i;
}

gocaml_int str_size(gocaml_string const s)
{
    return (gocaml_int) s.size;
}

gocaml_bool __str_equal(gocaml_string const l, gocaml_string const r)
{
    if (l.size != r.size) {
        return (gocaml_bool) 0;
    }
    return (gocaml_bool)(strcmp((char *)l.chars, (char *)r.chars) == 0);
}

