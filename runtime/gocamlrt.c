#include <stdio.h>
#include <inttypes.h>
#include <stdlib.h>
#include <string.h>
#include <gc.h>
#include "gocaml.h"

#define SNPRINTF_MAX 128
#define LINE_MAX 1024

extern int __gocaml_main();

// string array for argv
typedef struct {
    gocaml_string *buf;
    gocaml_int size;
} argv_t;
argv_t argv;

int main(int const argc, char const* const argv_[]) {
    GC_init();
    gocaml_string *ptr = (gocaml_string *) GC_malloc(argc * sizeof(gocaml_string *));
    for (int i = 0; i < argc; ++i) {
        gocaml_string s;
        s.chars = (int8_t *) argv_[i];
        s.size = strlen(argv_[i]);
        *(ptr + i) = s;
    }
    argv.buf = ptr;
    argv.size = (int64_t) argc;
    return __gocaml_main();
}

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

// Do not expect Nul-terminated string because of string slices
void print_str(gocaml_string const s)
{
    printf("%.*s", (int) s.size, (char *)s.chars);
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
    printf("%.*s\n", (int) s.size, (char *)s.chars);
}

gocaml_int float_to_int(gocaml_float const f)
{
    return (gocaml_int) f;
}

gocaml_float int_to_float(gocaml_int const i)
{
    return (gocaml_float) i;
}

gocaml_int str_length(gocaml_string const s)
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

gocaml_string str_concat(gocaml_string const l, gocaml_string const r)
{
    size_t const new_size = l.size + r.size + 1;
    char *const new_ptr = (char *) GC_malloc(new_size);

    strncpy(new_ptr, (char *) l.chars, (size_t) l.size);
    strncpy(new_ptr + l.size, (char *) r.chars, (size_t) r.size);
    new_ptr[new_size - 1] = '\0';

    gocaml_string ret;
    ret.chars = (int8_t *) new_ptr;
    ret.size = (gocaml_int) new_size;
    return ret;
}

// Slice [start,last) like Go's str[start:last]
gocaml_string substr(gocaml_string const s, gocaml_int const start, gocaml_int const last)
{
    if (s.size == 0) {
        return s;
    }

    int64_t start_idx = start;
    if (s.size <= start_idx) {
        start_idx = s.size; // This makes empty string
    } else if (start_idx < 0) {
        start_idx = 0;
    }

    int64_t last_idx = last;
    if (last_idx < 0) {
        last_idx = 0;
    } else if (s.size <= last_idx) {
        last_idx = s.size;
    }

    int64_t new_size = last_idx - start_idx;
    if (new_size < 0) {
        new_size = 0;
    }

    int8_t *const new_ptr = s.chars + start_idx;
    gocaml_string ret;
    ret.chars = new_ptr;
    ret.size = new_size;
    return ret;
}

gocaml_string int_to_str(gocaml_int const i)
{
    char *const s = GC_malloc(SNPRINTF_MAX);
    int const n = snprintf(s, SNPRINTF_MAX, "%" PRId64, i);
    gocaml_string ret;
    ret.chars = (int8_t *) s;
    ret.size = (int64_t) n;
    return ret;
}

gocaml_string float_to_str(gocaml_float const f)
{
    char *s = GC_malloc(SNPRINTF_MAX);
    int const n = snprintf(s, SNPRINTF_MAX, "%lg", f);
    gocaml_string ret;
    ret.chars = (int8_t *) s;
    ret.size = (int64_t) n;
    return ret;
}

gocaml_int str_to_int(gocaml_string const s)
{
    char const backup = s.chars[s.size];

    // Note:
    // Need to guard with this 'if' statement because when the string is allocated as global
    // constant variable, we can't modify it. And we does not need to modify global constant
    // string because it is always NUL-terminated.
    if (s.chars[s.size] != '\0') {
        s.chars[s.size] = '\0'; // Ensure to terminate with NUL.
    }

    int const i = atoi((char *) s.chars);

    if (s.chars[s.size] != '\0') {
        s.chars[s.size] = backup;
    }

    return (gocaml_int) i;
}

gocaml_float str_to_float(gocaml_string const s)
{
    char const backup = s.chars[s.size];

    // Note:
    // Need to guard with this 'if' statement because when the string is allocated as global
    // constant variable, we can't modify it. And we does not need to modify global constant
    // string because it is always NUL-terminated.
    if (s.chars[s.size] != '\0') {
        s.chars[s.size] = '\0'; // Ensure to terminate with NUL.
    }

    double const f = atof((char *) s.chars);

    if (s.chars[s.size] != '\0') {
        s.chars[s.size] = backup;
    }

    return (gocaml_float) f;
}

gocaml_string get_line(gocaml_unit _)
{
    (void) _;
    char *const s = fgets((char *) GC_malloc(LINE_MAX), LINE_MAX, stdin);
    gocaml_string ret;

    if (s == NULL) {
        char *const emp = GC_malloc(1);
        emp[0] = '\0';
        ret.chars = (int8_t *) emp;
        ret.size = 0;
        return ret;
    }

    ret.chars = (int8_t *) s;
    ret.size = (gocaml_int) strlen(s);
    return ret;
}

gocaml_string get_char(gocaml_unit _)
{
    (void) _;
    gocaml_string ret;
    int *const s = (int *) GC_malloc(sizeof(int) * 2);
    *s = getchar();
    *(s + 1) = '\0';
    ret.chars = (int8_t *) s;
    ret.size = 1;
    return ret;
}
