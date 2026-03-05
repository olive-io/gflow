import dayjs from 'dayjs';

const toNumber = (value: number | string): number => (typeof value === 'string' ? Number(value) : value);

export const normalizeToMilliseconds = (value: number | string): number => {
  const ts = toNumber(value);
  if (!Number.isFinite(ts) || ts <= 0) {
    return 0;
  }
  if (ts > 100000000000) {
    return Math.floor(ts);
  }
  return Math.floor(ts * 1000);
};

export const normalizeToSeconds = (value: number | string): number => {
  const ts = toNumber(value);
  if (!Number.isFinite(ts) || ts <= 0) {
    return 0;
  }
  if (ts > 100000000000) {
    return Math.floor(ts / 1000);
  }
  return Math.floor(ts);
};

export const formatDate = (timestamp: number | string | undefined): string => {
  if (!timestamp) return '-';
  const ts = normalizeToMilliseconds(timestamp);
  if (ts === 0) return '-';
  return dayjs(ts).format('YYYY-MM-DD HH:mm:ss');
};

export const formatDateShort = (timestamp: number | string | undefined): string => {
  if (!timestamp) return '-';
  const ts = normalizeToMilliseconds(timestamp);
  if (ts === 0) return '-';
  return dayjs(ts).format('YYYY-MM-DD');
};

export const formatDuration = (startTime?: number | string, endTime?: number | string): string => {
  if (!startTime) return '-';
  const start = normalizeToSeconds(startTime);
  if (start === 0) return '-';

  let end: number;
  if (endTime) {
    end = normalizeToSeconds(endTime);
    if (end === 0) {
      end = Math.floor(Date.now() / 1000);
    }
  } else {
    end = Math.floor(Date.now() / 1000);
  }
  
  const duration = end - start;
  if (duration < 0) return '-';
  if (duration < 60) return `${duration}s`;
  if (duration < 3600) return `${Math.floor(duration / 60)}m ${Math.floor(duration % 60)}s`;
  return `${Math.floor(duration / 3600)}h ${Math.floor((duration % 3600) / 60)}m`;
};
