import { writable, type Readable, type Writable } from "svelte/store";
import { onMount } from "svelte";
import type { Result } from "@/api/api.utils";

const useFlag: (
  value: boolean,
  onSet?: () => void,
  onReset?: () => void,
) => [Writable<boolean>, () => void, () => void] = (
  value = false,
  onSet?,
  onReset?,
) => {
  let flag = writable(value);

  const enableFlag = () => {
    onSet?.();
    flag.set(true);
  };

  const disableFlag = () => {
    onReset?.();
    flag.set(false);
  };

  return [flag, enableFlag, disableFlag];
};

const useToggle: (
  value: boolean,
  onSet?: () => void,
  onReset?: () => void,
) => [Writable<boolean>, () => void] = (value = false, onSet?, onReset?) => {
  let t: boolean = value;

  const [toggle, set, reset] = useFlag(t, onSet, onReset);

  const flip = () => {
    t = !t;
    (t ? set : reset)();
  };

  return [toggle, flip];
}

interface QueryResult<T> {
  data?: Readable<T>,
  isLoading: Readable<boolean>,
  error?: Readable<string>,
};

const useQuery: <T>(
  query: (...args: any[]) => Promise<Result<T> | Result<undefined>>,
  ...args: any[]
) => QueryResult<T> = <T>(query: (...args: any[]) => Promise<Result<T> | Result<undefined>>, ...args: any[]) => {
  let data: Writable<T> = writable(null);
  let isLoading: Writable<boolean> = writable(true);
  let error: Writable<string> = writable(null);

  onMount(async () => {
    const { data: content, status, message } = await query(...args);
    
    if (status === "OK") {
      data.set(content);
    } else {
      error.set(message);
    }

    isLoading.set(false);
  });

  return {
    data: data,
    isLoading: isLoading,
    error: error,
  };
};

export {
  useFlag,
  useToggle,
  useQuery,
};