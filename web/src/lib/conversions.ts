type TransformerFunc<T> = (v: T) => T;

type StringTransformer = TransformerFunc<string>;

const nullTransformer: TransformerFunc<any> = (v: any) => v;

const snakeToCamel: StringTransformer = (v: string) => v.replace(/([_][a-z])/g, group => group.toUpperCase().replace('_', ''));
const camelToSnake: StringTransformer = (v: string) => v.replace(/([a-z][A-Z])/g, group => group.toLowerCase().split("").join("_"));

const transformKeys = (o: any, t: StringTransformer = nullTransformer): any => {
  if (typeof o !== 'object' || [null, undefined].includes(o)) {
    return o;
  }

  if (Array.isArray(o)) {
    return o.map(v => transformKeys(v, t));
  }

  return Object.fromEntries(Object.entries(o).map(([key, value]) => [t(key), transformKeys(value, t)]));
}

export {
  snakeToCamel,
  camelToSnake,
  transformKeys,
};