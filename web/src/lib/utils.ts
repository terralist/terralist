const defaultIfNull = (v: any, d: any) => { 
  return v === null ? d : v; 
}

const indent: (params: {
  s: string,
  n: number,
  reverse?: boolean,
  trim?: boolean,
}) => string = ({
  s, 
  n,
  reverse = false,
  trim = true,
}) => {
  const lines = s.split("\n");

  const trimmedLines = lines.slice(trim ? 1 : 0, trim ? lines.length - 1 : lines.length);

  const formattedLines = reverse ? 
    trimmedLines.map(v => v.slice(n)) : 
    trimmedLines.map(v => `${new Array(n + 1).join(" ")}${v}`);
  
  return formattedLines.join("\n");
};

const toSeconds = 1000;
const toMinutes = toSeconds * 60;
const toHours = toMinutes * 60;
const toDays = toHours * 24;
const toMonths = toDays * 31;
const toYears = toMonths * 12;

const timeSince = (since: Date, from: Date = new Date()): string => {
  const diff: number = from.getTime() - since.getTime();

  const formatter = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

  const convert = (v: number, b: number) => {
    return Math.floor(Math.round(v / b));
  };

  if (diff > toYears) {
    const yearsSince = convert(diff, toYears);
    return formatter.format(-yearsSince, 'year');
  }

  if (diff > toMonths) {
    const monthsSince = convert(diff, toMonths);
    return formatter.format(-monthsSince, 'month');
  }
  
  if (diff > toDays) {
    const daysSince = convert(diff, toDays);
    return formatter.format(-daysSince, 'day');
  }

  if (diff > toHours) {
    const hoursSince = convert(diff, toHours);
    return formatter.format(-hoursSince, 'hour');
  }
  
  if (diff > toMinutes) {
    const minutesSince = convert(diff, toMinutes);
    return formatter.format(-minutesSince, 'minute');
  }

  const secondsSince = convert(diff, toSeconds);
  return formatter.format(-secondsSince, 'second');
}

export {
  defaultIfNull,
  indent,
  timeSince
};