/**
 * Format file size in bytes to human-readable format
 * @param bytes - File size in bytes
 * @returns Formatted string (e.g., "1.5 MB", "500 KB")
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  if (bytes < 0) return 'Invalid';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  // Ensure we don't exceed array bounds
  const unitIndex = Math.min(i, units.length - 1);

  // Format with 1 decimal place for KB and above, no decimals for bytes
  if (unitIndex === 0) {
    return `${bytes} ${units[0]}`;
  }

  const value = bytes / Math.pow(k, unitIndex);
  return `${value.toFixed(1)} ${units[unitIndex]}`;
}
