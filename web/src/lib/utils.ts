const defaultIfNull = (v: any, d: any) => { 
  return v === null ? d : v; 
}

export {
  defaultIfNull
};