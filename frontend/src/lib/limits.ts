/** Shared upload / compose limits (keep in sync with FileUploader + BFF). */
export const MAX_FILE_BYTES = 2 * 1024 * 1024; // 2 MiB raw
export const MAX_FILES = 4;
export const MAX_TITLE_CHARS = 255;
export const MAX_TEXT_CHARS = 16_000;
/** Overall request body cap under adapter BODY_SIZE_LIMIT (16M). */
export const MAX_JSON_BODY_CHARS = 12 * 1024 * 1024;
