#ifndef FLUIDAUDIO_H
#define FLUIDAUDIO_H

#include <stdint.h>

// --- Lifecycle ---
void *fluidaudio_bridge_create(void);
void fluidaudio_bridge_destroy(void *ptr);
void fluidaudio_cleanup(void *ptr);
void fluidaudio_free_string(char *s);

// --- ASR ---
int32_t fluidaudio_initialize_asr(void *ptr);
int32_t fluidaudio_transcribe_file(void *ptr, const char *path, char **out_text,
                                   float *out_confidence, double *out_duration,
                                   double *out_processing_time,
                                   float *out_rtfx);
int32_t fluidaudio_transcribe_samples(void *ptr, const float *samples,
                                      uint32_t count, char **out_text,
                                      float *out_confidence,
                                      double *out_duration,
                                      double *out_processing_time,
                                      float *out_rtfx);
int32_t fluidaudio_is_asr_available(void *ptr);

// --- Streaming ASR ---
int32_t fluidaudio_initialize_streaming_asr(void *ptr);
int32_t fluidaudio_streaming_asr_start(void *ptr);
int32_t fluidaudio_streaming_asr_feed(void *ptr, const float *samples,
                                      uint32_t count);
int32_t fluidaudio_streaming_asr_finish(void *ptr, char **out_text);
int32_t fluidaudio_transcribe_file_streaming(
    void *ptr, const char *path, char **out_text, float *out_confidence,
    double *out_duration, double *out_processing_time, float *out_rtfx);
int32_t fluidaudio_is_streaming_asr_available(void *ptr);

// --- VAD ---
int32_t fluidaudio_initialize_vad(void *ptr, float threshold);
int32_t fluidaudio_is_vad_available(void *ptr);

// --- Diarization ---
int32_t fluidaudio_initialize_diarization(void *ptr, double threshold,
                                          int32_t min_speakers,
                                          int32_t max_speakers);
int32_t fluidaudio_diarize_file(void *ptr, const char *path,
                                char ***out_speaker_ids,
                                float **out_start_times, float **out_end_times,
                                float **out_quality_scores,
                                uint32_t *out_count);
int32_t fluidaudio_is_diarization_available(void *ptr);
void fluidaudio_free_diarization_result(char **speaker_ids, float *start_times,
                                        float *end_times, float *quality_scores,
                                        uint32_t count);

// --- Qwen3 ASR ---
int32_t fluidaudio_initialize_qwen3_asr(void *ptr);
int32_t fluidaudio_qwen3_transcribe_file(void *ptr, const char *path,
                                         const char *language, char **out_text,
                                         float *out_confidence,
                                         double *out_duration,
                                         double *out_processing_time,
                                         float *out_rtfx);
int32_t fluidaudio_qwen3_transcribe_samples(
    void *ptr, const float *samples, uint32_t count, const char *language,
    char **out_text, float *out_confidence, double *out_duration,
    double *out_processing_time, float *out_rtfx);
int32_t fluidaudio_is_qwen3_asr_available(void *ptr);

// --- Qwen3 Streaming ---
int32_t fluidaudio_initialize_qwen3_streaming(void *ptr);
int32_t fluidaudio_qwen3_streaming_start(void *ptr, const char *language,
                                         double min_audio_seconds,
                                         double chunk_seconds,
                                         double max_audio_seconds);
int32_t fluidaudio_qwen3_streaming_feed(void *ptr, const float *samples,
                                        uint32_t count,
                                        char **out_partial_text);
int32_t fluidaudio_qwen3_streaming_finish(void *ptr, char **out_text);
int32_t fluidaudio_is_qwen3_streaming_available(void *ptr);

// --- System Info ---
void fluidaudio_get_platform(char **out);
void fluidaudio_get_chip_name(char **out);
double fluidaudio_get_memory_gb(void);
int32_t fluidaudio_is_apple_silicon(void);

#endif
