type TransformerFunc<T> = (v: T) => T;

type StringTransformer = TransformerFunc<string>;

const snakeToCamel: StringTransformer = (v: string) =>
  v.replace(/([_][a-z])/g, group => group.toUpperCase().replace('_', ''));
const camelToSnake: StringTransformer = (v: string) =>
  v.replace(/([a-z][A-Z])/g, group => group.toLowerCase().split('').join('_'));

const transformKeys = <T>(o: T, t?: StringTransformer): T => {
  if (typeof o !== 'object' || o == null || o == undefined) {
    return o;
  }

  if (Array.isArray(o)) {
    return o.map(v => transformKeys(v, t)) as T;
  }

  t = t ?? ((v: string) => v);

  return Object.fromEntries(
    Object.entries(o).map(([key, value]) => [t(key), transformKeys(value, t)])
  ) as T;
};

export { snakeToCamel, camelToSnake, transformKeys };
