import { writable, type Writable } from "svelte/store";

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

export {
  useFlag,
  useToggle,
};