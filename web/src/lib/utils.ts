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

export {
  defaultIfNull,
  indent
};