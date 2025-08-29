/*
Copyright 2018 Embedded Microprocessor Benchmark Consortium (EEMBC)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Original Author: Shay Gal-on
*/

#ifndef CORE_MARK_H
#define CORE_MARK_H

/* Configuration */
#ifndef ITERATIONS
#define ITERATIONS 50000
#endif

#ifndef MAIN_HAS_NOARGC
#define MAIN_HAS_NOARGC 0
#endif

#ifndef MAIN_HAS_NORETURN
#define MAIN_HAS_NORETURN 0
#endif

#ifndef HAS_FLOAT
#define HAS_FLOAT 1
#endif

#ifndef CORE_TICKS
#define NSECS_PER_SEC CLOCKS_PER_SEC
#define EE_TICKS_PER_SEC (NSECS_PER_SEC / 1000)
#define GETMYTIME(_t) (*_t = clock())
#define MYTIMEDIFF(fin, ini) ((fin) - (ini))
#define TIMER_RES_DIVIDER 1
#define SAMPLE_TIME_IMPLEMENTATION 1
#define CLOCKS_PER_SEC 1000000
typedef clock_t CORE_TICKS;
#endif

#ifndef COMPILER_VERSION
#ifdef __GNUC__
#define COMPILER_VERSION "GCC"__VERSION__
#else
#define COMPILER_VERSION "Unknown"
#endif
#endif

#ifndef MEM_LOCATION
#define MEM_LOCATION "STACK"
#endif

/* Data types */
typedef signed short ee_s16;
typedef unsigned short ee_u16;
typedef signed int ee_s32;
typedef double ee_f32;
typedef unsigned char ee_u8;
typedef unsigned int ee_u32;
typedef ee_u32 ee_ptr_int;
typedef size_t ee_size_t;

#define CORETIMETYPE ee_u32
typedef ee_u16 MATRES;
typedef ee_s16 MATDAT;

/* Actual benchmark execution in ticks */
#define SEED1_VAL 0x3415
#define SEED2_VAL 0x3415
#define SEED3_VAL 0x66

#define MEM_METHOD_MALLOC 0
#define MEM_METHOD_STATIC 1
#define MEM_METHOD_STACK 2

#ifndef MEM_METHOD
#define MEM_METHOD MEM_METHOD_MALLOC
#endif

/* List data structures */
typedef struct list_data_s {
    ee_s16 data16;
    ee_s16 idx;
} list_data;

typedef struct list_head_s {
    struct list_head_s *next;
    struct list_data_s *info;
} list_head;

/*Matrix parameters*/
#define MATDAT_INT 1

typedef struct MAT_PARAMS_S {
    int N;
    MATDAT *A;
    MATDAT *B;
    MATRES *C;
} mat_params;

/* State machine */
typedef struct RESULTS_S {
    ee_s16   seed1;    /* Initializing seed */
    ee_s16   seed2;    /* Initializing seed */
    ee_s16   seed3;    /* Initializing seed */
    void     *memblock[4]; /* Pointers to safe memory blocks */
    ee_u32   size;     /* Size of the data */
    ee_u32   iterations; /* Number of iterations to run */
    ee_u32   execs;    /* Bitmask of operations to execute */
    struct list_head_s *list;
    mat_params mat;
    ee_u16   crc;
    ee_u16   crclist;
    ee_u16   crcmatrix;
    ee_u16   crcstate;
    ee_s16   err;
    CORE_TICKS *port;
} core_results;

/* Function prototypes */
int main(void);
void start_time(void);
void stop_time(void);
CORE_TICKS get_time(void);
ee_u32 default_num_contexts(void);

/* benchmark functions */
CORETIMETYPE iterate(void *pres);
void *portable_malloc(ee_size_t size);
void portable_free(void *p);
ee_s32 get_seed_32(int i);
ee_u16 crcu8(ee_u8 data, ee_u16 crc);
ee_u16 crc16(ee_s16 newval, ee_u16 crc);
ee_u16 crcu16(ee_u16 newval, ee_u16 crc);
ee_u16 crcu32(ee_u32 newval, ee_u16 crc);
void *iterate_matrix(void *pres);
ee_s32 matrix_test(ee_u32 N, MATRES *C, MATDAT *A, MATDAT *B, MATDAT val);
ee_s32 matrix_sum(ee_u32 N, MATRES *C, MATDAT clipval);
void matrix_mul_const(ee_u32 N, MATRES *C, MATDAT *A, MATDAT val);
void matrix_mul_vect(ee_u32 N, MATRES *C, MATDAT *A, MATDAT *B);
void matrix_mul_matrix(ee_u32 N, MATRES *C, MATDAT *A, MATDAT *B);
void matrix_mul_matrix_bitextract(ee_u32 N, MATRES *C, MATDAT *A, MATDAT *B);
void *iterate_list(void *pres);
list_head *core_list_init(ee_u32 blksize, list_head *memblock, ee_s16 seed);
ee_s16 core_list_undo_remove(list_head *item_removed, list_head *item_modified, list_head *memblock);
list_head *core_list_remove(list_head *item);
list_head *core_list_find(list_head *list, list_data *info);
list_head *core_list_reverse(list_head *list);
list_head *core_list_mergesort(list_head *list, core_results *res);
ee_s16 calc_func(ee_s16 *pdata, core_results *res);
ee_s16 cmp_complex(ee_s16 val1, ee_s16 val2, core_results *res);
ee_s16 cmp_idx(ee_s16 val1, ee_s16 val2, core_results *res);
void *iterate_state(void *pres);
ee_u32 core_bench_state(ee_u32 blksize, ee_u8 *memblock, ee_s16 seed1, ee_s16 seed2, ee_s16 step, ee_u16 crc);
enum CORE_STATE { CORE_START=0, CORE_INVALID, CORE_S1, CORE_S2, CORE_INT, CORE_FLOAT, CORE_EXPONENT, CORE_SCIENTIFIC };
typedef struct CORE_STATE_S {
    ee_s32 x;
    ee_s32 y;
    ee_s32 info;
    ee_u16 crc;
} core_state_s;

#ifndef MAIN_HAS_NOARGC
int ee_main(void);
#else
int ee_main(int argc, char *argv[]);
#endif

#endif /* CORE_MARK_H */