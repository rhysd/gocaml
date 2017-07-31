#include <stdio.h>
#include <inttypes.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <math.h>
#include <gc.h>
#include "gocaml.h"

#define SNPRINTF_MAX 128
#define LINE_MAX 1024
#define BUF_CHUNK 1024

// Note:
// Need to guard with this 'if' statement because when the string is allocated as global
// constant variable, we can't modify it. And we does not need to modify global constant
// string because it is always NUL-terminated.
#define GOCAML_STRING_ENSURE_NULL(x) \
    char const backup_null_char_ ## x = (x).chars[(x).size]; \
    if ((x).chars[(x).size] != '\0') { \
        (x).chars[(x).size] = '\0'; \
    }

#define GOCAML_STRING_RESTORE_NULL(x) \
    if ((x).chars[(x).size] != '\0') { \
        (x).chars[(x).size] = backup_null_char_ ## x; \
    }

extern int __gocaml_main();

// Constants
double gocaml_infinity = INFINITY;
double gocaml_nan = NAN;

// string array for argv
typedef struct {
    gocaml_string *buf;
    gocaml_int size;
} argv_t;
argv_t argv;

typedef struct {
    gocaml_float fst;
    gocaml_float snd;
} ff_pair_t;

typedef struct {
    gocaml_float fst;
    gocaml_int snd;
} fi_pair_t;

typedef struct {
    gocaml_int fst;
    gocaml_float snd;
} if_pair_t;

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
    int const cmp = strncmp((char *)l.chars, (char *)r.chars, (size_t) l.size);
    return (gocaml_bool) cmp == 0;
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
gocaml_string str_sub(gocaml_string const s, gocaml_int const start, gocaml_int const last)
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
    GOCAML_STRING_ENSURE_NULL(s);

    int const i = atoi((char *) s.chars);

    GOCAML_STRING_RESTORE_NULL(s);

    return (gocaml_int) i;
}

gocaml_float str_to_float(gocaml_string const s)
{
    GOCAML_STRING_ENSURE_NULL(s);

    double const f = atof((char *) s.chars);

    GOCAML_STRING_RESTORE_NULL(s);

    return (gocaml_float) f;
}

gocaml_string get_line(gocaml_unit _)
{
    (void) _;
    char *const s = fgets((char *) GC_malloc(sizeof(char) * LINE_MAX), LINE_MAX, stdin);
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

gocaml_int to_char_code(gocaml_string const s)
{
    if (s.size == 0) {
        return 0;
    }
    return (int64_t) s.chars[0];
}

gocaml_string from_char_code(gocaml_int const i)
{
    char *const ptr = GC_malloc(2);
    *ptr = (char) i;
    *(ptr + 1) = '\0';
    gocaml_string ret;
    ret.chars = (int8_t *) ptr;
    ret.size = 1;
    return ret;
}

void do_garbage_collection(gocaml_unit _)
{
    (void) _;
    GC_gcollect();
}

void enable_garbage_collection(gocaml_unit _)
{
    (void) _;
    GC_enable();
}

void disable_garbage_collection(gocaml_unit _)
{
    (void) _;
    GC_disable();
}

gocaml_int bit_and(gocaml_int const l, gocaml_int const r)
{
    return l & r;
}
gocaml_int bit_or(gocaml_int const l, gocaml_int const r)
{
    return l | r;
}
gocaml_int bit_xor(gocaml_int const l, gocaml_int const r)
{
    return l ^ r;
}
gocaml_int bit_rsft(gocaml_int const l, gocaml_int const r)
{
    return l >> r;
}
gocaml_int bit_lsft(gocaml_int const l, gocaml_int const r)
{
    return l << r;
}
gocaml_int bit_inv(gocaml_int const i)
{
    return ~i;
}

ff_pair_t *gocaml_modf(gocaml_float const f)
{
    double fractional, integral;
    ff_pair_t *ret;

    fractional = modf(f, &integral);
    ret = (ff_pair_t *) GC_malloc(sizeof(ff_pair_t));
    ret->fst = fractional;
    ret->snd = integral;
    return ret;
}

fi_pair_t *gocaml_frexp(gocaml_float const f)
{
    double frac;
    int exp;
    fi_pair_t *ret;

    frac = frexp(f, &exp);
    ret = (fi_pair_t *) GC_malloc(sizeof(fi_pair_t));
    ret->fst = frac;
    ret->snd = exp;
    return ret;
}

gocaml_float gocaml_ldexp(gocaml_float const f, gocaml_int const i)
{
    return ldexp(f, (int) i);
}

gocaml_int time_now(gocaml_unit _)
{
    (void) _;
    return (gocaml_int) time(NULL);
}

gocaml_string read_file(gocaml_string const filename)
{
    GOCAML_STRING_ENSURE_NULL(filename);

    FILE *file = fopen((char *) filename.chars, "r");
    if (file == NULL) {
        GOCAML_STRING_RESTORE_NULL(filename);
        gocaml_string none;
        none.chars = NULL;
        return none;
    }

    char c;
    int idx = 0;
    int num_chunk = 1;
    char *buf = (char *) GC_malloc(sizeof(char) * BUF_CHUNK * num_chunk);
    while ((c = getc(file)) != EOF) {
        buf[idx] = c;
        if ((BUF_CHUNK * num_chunk - 2) <= idx) {
            char *old = buf;
            num_chunk++;
            buf = (char *) GC_malloc(sizeof(char) * BUF_CHUNK * num_chunk);
            memcpy(buf, old, sizeof(char) * (num_chunk - 1));
            GC_free(old);
        }
        idx++;
    }
    buf[idx] = '\0';
    fclose(file);

    gocaml_string ret;
    ret.chars = (int8_t *)buf;
    ret.size = (gocaml_int) idx;
    GOCAML_STRING_RESTORE_NULL(filename);
    return ret;
}

gocaml_bool write_file(gocaml_string const filename, gocaml_string const content)
{
    GOCAML_STRING_ENSURE_NULL(filename);
    GOCAML_STRING_ENSURE_NULL(content);

    FILE *file = fopen((char *) filename.chars, "w");
    if (file == NULL) {
        GOCAML_STRING_RESTORE_NULL(filename);
        GOCAML_STRING_RESTORE_NULL(content);
        return (gocaml_bool) 0;
    }

    for (int i = 0; i < content.size; i++) {
        fputc(content.chars[i], file);
    }

    fclose(file);
    GOCAML_STRING_RESTORE_NULL(filename);
    GOCAML_STRING_RESTORE_NULL(content);
    return (gocaml_bool) 1;
}
