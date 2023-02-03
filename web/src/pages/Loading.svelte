<script lang="ts">
  import Navbar from "@/components/Navbar.svelte";
</script>

<Navbar />
<main class="mt-36 lg:mt-20 mx-10">
  <div class="absolute top-0 left-0 flex justify-center items-center text-center w-screen h-screen -z-10">
    <svg
      xmlns="http://www.w3.org/2000/svg"
      xmlns:xlink="http://www.w3.org/1999/xlink"
      style="margin: auto; background: none; display: block; shape-rendering: auto;"
      width="8rem"
      height="8rem"
      viewBox="0 0 100 100"
      preserveAspectRatio="xMidYMid"
    >
      <g
        transform="translate(50 50) scale(0.7000000000000001) translate(-50 -50)"
      >
        <g>
          <animateTransform
            attributeName="transform"
            type="rotate"
            repeatCount="indefinite"
            calcMode="spline"
            dur="4s"
            values="0 50 50;90 50 50;180 50 50;270 50 50;360 50 50"
            keyTimes="0;0.25;0.5;0.75;1"
            keySplines="0 1 0 1;0 1 0 1;0 1 0 1;0 1 0 1"
          />
          <g>
            <animateTransform
              attributeName="transform"
              type="scale"
              dur="1s"
              repeatCount="indefinite"
              calcMode="spline"
              values="1;1;0.5"
              keyTimes="0;0.5;1"
              keySplines="1 0 0 1;1 0 0 1"
            />
            <g transform="translate(25 25)">
              <rect x="-25" y="-25" width="52" height="52" fill="#55c89d">
                <animate
                  attributeName="fill"
                  dur="4s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="#55c89d;#309d8e;#247074;#204651;#55c89d"
                  keyTimes="0;0.25;0.5;0.75;1"
                  keySplines="0 1 0 1;0 1 0 1;0 1 0 1;0 1 0 1"
                />
              </rect>
            </g>
            <g transform="translate(25 75)">
              <rect x="-25" y="-25" width="52" height="50" fill="#55c89d">
                <animateTransform
                  attributeName="transform"
                  type="scale"
                  dur="1s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="0;1;1"
                  keyTimes="0;0.5;1"
                  keySplines="1 0 0 1;1 0 0 1"
                />
                <animate
                  attributeName="fill"
                  dur="4s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="#55c89d;#309d8e;#247074;#204651;#55c89d"
                  keyTimes="0;0.25;0.5;0.75;1"
                  keySplines="0 1 0 1;0 1 0 1;0 1 0 1;0 1 0 1"
                />
              </rect>
            </g>
            <g transform="translate(75 25)">
              <rect x="-25" y="-25" width="50" height="52" fill="#55c89d">
                <animateTransform
                  attributeName="transform"
                  type="scale"
                  dur="1s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="0;1;1"
                  keyTimes="0;0.5;1"
                  keySplines="1 0 0 1;1 0 0 1"
                />
                <animate
                  attributeName="fill"
                  dur="4s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="#55c89d;#309d8e;#247074;#204651;#55c89d"
                  keyTimes="0;0.25;0.5;0.75;1"
                  keySplines="0 1 0 1;0 1 0 1;0 1 0 1;0 1 0 1"
                />
              </rect>
            </g>
            <g transform="translate(75 75)">
              <rect x="-25" y="-25" width="50" height="50" fill="#55c89d">
                <animateTransform
                  attributeName="transform"
                  type="scale"
                  dur="1s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="0;1;1"
                  keyTimes="0;0.5;1"
                  keySplines="1 0 0 1;1 0 0 1"
                />
                <animate
                  attributeName="fill"
                  dur="4s"
                  repeatCount="indefinite"
                  calcMode="spline"
                  values="#55c89d;#309d8e;#247074;#204651;#55c89d"
                  keyTimes="0;0.25;0.5;0.75;1"
                  keySplines="0 1 0 1;0 1 0 1;0 1 0 1;0 1 0 1"
                />
              </rect>
            </g>
          </g>
        </g>
      </g>
    </svg>
  </div>
</main>
