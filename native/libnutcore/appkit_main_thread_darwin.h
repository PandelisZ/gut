#pragma once

#ifdef __cplusplus
extern "C" {
#endif

typedef void (*gut_main_thread_callback)(void *context);

void gut_run_on_main_thread_sync(gut_main_thread_callback callback, void *context);
void gut_run_on_main_thread_async(gut_main_thread_callback callback, void *context);

#ifdef __cplusplus
}
#endif
