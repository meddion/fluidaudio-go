#ifndef FLUIDAUDIO_H
#define FLUIDAUDIO_H

#include <stdint.h>

// --- Lifecycle ---
void *fluidaudio_diarizer_create(void);
void fluidaudio_diarizer_destroy(void *ptr);
void fluidaudio_cleanup(void *ptr);
void fluidaudio_free_string(char *s);

// --- Diarization ---
int32_t fluidaudio_initialize_diarization(
    void *ptr, float onset_threshold, float offset_threshold,
    int32_t onset_pad_frames, int32_t offset_pad_frames, int32_t min_frames_on,
    int32_t min_frames_off, int32_t compute, int32_t variant);
int32_t fluidaudio_diarize_offline(void *ptr, const char *path,
                                   int32_t **out_speaker_ids,
                                   float **out_start_times,
                                   float **out_end_times, uint32_t *out_count);

void fluidaudio_free_diarize_offline(int32_t *speaker_ids, float *start_times,
                                     float *end_times);

// --- Streaming Diarization ---
int32_t fluidaudio_diarize_process_audio(
    void *ptr, const float *samples, uint32_t sample_count,
    double source_sample_rate, int32_t **out_speaker_ids,
    float **out_start_times, float **out_end_times, uint32_t *out_count);
int32_t fluidaudio_diarize_finalize(void *ptr, int32_t **out_speaker_ids,
                                    float **out_start_times,
                                    float **out_end_times, uint32_t *out_count);
void fluidaudio_free_segments(int32_t *speaker_ids, float *start_times,
                              float *end_times);

// --- System Info ---
void fluidaudio_get_platform(char **out);
void fluidaudio_get_chip_name(char **out);
double fluidaudio_get_memory_gb(void);
int32_t fluidaudio_is_apple_silicon(void);

#endif
