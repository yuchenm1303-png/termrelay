<template>
  <div class="cosmos" aria-hidden="true">
    <div class="cosmos-sky"></div>
    <div class="cosmos-horizon"></div>
    <div class="cosmos-nebula nebula-cyan"></div>
    <div class="cosmos-nebula nebula-violet"></div>
    <div class="cosmos-nebula nebula-rose"></div>

    <div class="star-dust dust-a"></div>
    <div class="star-dust dust-b"></div>
    <div class="star-dust dust-c"></div>

    <span
      v-for="star in stars"
      :key="star.id"
      class="star"
      :class="[`star-${star.kind}`, `star-${star.color}`]"
      :style="starStyle(star)"
    ></span>

    <div class="orbit">
      <span class="orbit-ring"></span>
      <span class="orbit-core"></span>
    </div>

    <div class="cloud-bank cloud-bank-left"></div>
    <div class="cloud-bank cloud-bank-right"></div>
    <div class="cosmos-vignette"></div>
    <div class="cosmos-grain"></div>
  </div>
</template>

<script setup lang="ts">
type StarKind = 'pin' | 'soft' | 'flare' | 'diamond'
type StarColor = 'ivory' | 'cyan' | 'rose'

type Star = {
  id: number
  x: number
  y: number
  size: number
  kind: StarKind
  color: StarColor
  delay: number
  duration: number
  opacity: number
}

const stars: Star[] = [
  { id: 1, x: 5.5, y: 36, size: 1.1, kind: 'flare', color: 'cyan', delay: -1.4, duration: 6.2, opacity: 0.95 },
  { id: 2, x: 13, y: 48, size: 0.75, kind: 'diamond', color: 'ivory', delay: -2.8, duration: 7.5, opacity: 0.78 },
  { id: 3, x: 18.5, y: 29, size: 0.48, kind: 'pin', color: 'ivory', delay: -1.2, duration: 5.6, opacity: 0.68 },
  { id: 4, x: 25, y: 43, size: 0.8, kind: 'flare', color: 'cyan', delay: -4.1, duration: 8.4, opacity: 0.82 },
  { id: 5, x: 31, y: 24, size: 0.38, kind: 'pin', color: 'rose', delay: -0.8, duration: 6.7, opacity: 0.52 },
  { id: 6, x: 38.5, y: 51, size: 0.7, kind: 'diamond', color: 'rose', delay: -3.3, duration: 7.2, opacity: 0.64 },
  { id: 7, x: 46, y: 34, size: 0.52, kind: 'soft', color: 'ivory', delay: -5.1, duration: 9.1, opacity: 0.7 },
  { id: 8, x: 51, y: 43, size: 0.95, kind: 'flare', color: 'cyan', delay: -2.2, duration: 6.9, opacity: 0.96 },
  { id: 9, x: 59, y: 29, size: 1.05, kind: 'flare', color: 'ivory', delay: -3.8, duration: 8.8, opacity: 0.9 },
  { id: 10, x: 64, y: 51, size: 0.42, kind: 'pin', color: 'ivory', delay: -1.6, duration: 6.4, opacity: 0.58 },
  { id: 11, x: 71, y: 39, size: 0.65, kind: 'diamond', color: 'cyan', delay: -4.7, duration: 7.8, opacity: 0.75 },
  { id: 12, x: 78, y: 26, size: 0.34, kind: 'pin', color: 'rose', delay: -2.4, duration: 5.9, opacity: 0.47 },
  { id: 13, x: 87.5, y: 31, size: 1.15, kind: 'flare', color: 'cyan', delay: -5.5, duration: 9.4, opacity: 0.98 },
  { id: 14, x: 91, y: 49, size: 0.54, kind: 'soft', color: 'ivory', delay: -3.1, duration: 7.1, opacity: 0.65 },
  { id: 15, x: 96, y: 41, size: 0.32, kind: 'pin', color: 'ivory', delay: -0.4, duration: 6.5, opacity: 0.5 },
  { id: 16, x: 9, y: 61, size: 0.42, kind: 'pin', color: 'rose', delay: -4.2, duration: 8.1, opacity: 0.48 },
  { id: 17, x: 18, y: 67, size: 0.7, kind: 'flare', color: 'cyan', delay: -1.9, duration: 7.6, opacity: 0.72 },
  { id: 18, x: 34, y: 63, size: 0.36, kind: 'pin', color: 'ivory', delay: -2.7, duration: 6.1, opacity: 0.5 },
  { id: 19, x: 57, y: 68, size: 0.58, kind: 'diamond', color: 'rose', delay: -3.6, duration: 8.6, opacity: 0.56 },
  { id: 20, x: 74, y: 63, size: 0.44, kind: 'pin', color: 'cyan', delay: -1.1, duration: 6.8, opacity: 0.58 },
  { id: 21, x: 83, y: 72, size: 0.78, kind: 'flare', color: 'ivory', delay: -4.9, duration: 9.2, opacity: 0.72 },
  { id: 22, x: 94, y: 61, size: 0.45, kind: 'soft', color: 'rose', delay: -2.6, duration: 7.4, opacity: 0.55 }
]

function starStyle(star: Star) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    '--star-size': `${star.size}rem`,
    '--star-delay': `${star.delay}s`,
    '--star-duration': `${star.duration}s`,
    '--star-opacity': String(star.opacity)
  }
}
</script>

<style scoped>
.cosmos {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  isolation: isolate;
  background: #010207;
}

.cosmos-sky,
.cosmos-horizon,
.cosmos-vignette,
.cosmos-grain,
.star-dust,
.cosmos-nebula,
.cloud-bank {
  position: absolute;
  inset: 0;
}

.cosmos-sky {
  z-index: 0;
  background:
    radial-gradient(ellipse at 50% 56%, rgba(5, 48, 55, 0.48), transparent 38%),
    radial-gradient(ellipse at 34% 76%, rgba(116, 57, 119, 0.34), transparent 33%),
    radial-gradient(ellipse at 67% 79%, rgba(54, 55, 117, 0.38), transparent 34%),
    linear-gradient(
      180deg,
      #010205 0%,
      #010308 17%,
      #02070b 31%,
      #06232a 52%,
      #173444 64%,
      #4c315d 78%,
      #151b32 100%
    );
}

.cosmos-horizon {
  z-index: 1;
  top: 39%;
  bottom: auto;
  height: 48%;
  opacity: 0.78;
  background:
    radial-gradient(ellipse at 50% 31%, rgba(120, 249, 238, 0.16), transparent 42%),
    radial-gradient(ellipse at 25% 62%, rgba(255, 122, 187, 0.16), transparent 39%),
    radial-gradient(ellipse at 75% 67%, rgba(133, 105, 255, 0.17), transparent 40%);
  filter: blur(24px) saturate(115%);
}

.cosmos-nebula {
  z-index: 2;
  border-radius: 48% 52% 58% 42% / 61% 43% 57% 39%;
  filter: blur(78px) saturate(132%);
  mix-blend-mode: screen;
  opacity: 0.28;
}

.nebula-cyan {
  inset: 28% 48% 7% -20%;
  background:
    radial-gradient(ellipse at 48% 48%, rgba(55, 226, 218, 0.58), rgba(18, 89, 110, 0.21) 49%, transparent 74%);
  animation: nebula-cyan-drift 24s ease-in-out infinite alternate;
}

.nebula-violet {
  inset: 35% -18% 3% 50%;
  background:
    radial-gradient(ellipse at 46% 52%, rgba(144, 104, 233, 0.55), rgba(56, 53, 123, 0.22) 52%, transparent 76%);
  animation: nebula-violet-drift 27s ease-in-out infinite alternate;
}

.nebula-rose {
  inset: 57% 15% -22% 12%;
  opacity: 0.32;
  background:
    radial-gradient(ellipse at 50% 38%, rgba(243, 118, 178, 0.58), rgba(105, 51, 120, 0.19) 52%, transparent 77%);
  animation: nebula-rose-breathe 20s ease-in-out infinite alternate;
}

.star-dust {
  z-index: 3;
  background-repeat: repeat;
  animation: dust-drift 120s linear infinite;
}

.dust-a {
  opacity: 0.42;
  background-image:
    radial-gradient(circle, rgba(255, 255, 255, 0.92) 0 0.65px, transparent 1.2px),
    radial-gradient(circle, rgba(120, 255, 245, 0.72) 0 0.8px, transparent 1.35px);
  background-position: 21px 37px, 97px 173px;
  background-size: 168px 168px, 257px 257px;
}

.dust-b {
  opacity: 0.26;
  background-image:
    radial-gradient(circle, rgba(255, 186, 225, 0.9) 0 0.55px, transparent 1.15px),
    radial-gradient(circle, rgba(255, 255, 255, 0.82) 0 0.5px, transparent 1.05px);
  background-position: 83px 131px, 171px 43px;
  background-size: 341px 341px, 219px 219px;
  animation-direction: reverse;
  animation-duration: 150s;
}

.dust-c {
  opacity: 0.18;
  background-image:
    radial-gradient(circle, rgba(170, 151, 255, 0.86) 0 0.48px, transparent 1px);
  background-position: 19px 91px;
  background-size: 129px 129px;
  animation-duration: 95s;
}

.star {
  --star-size: 0.5rem;
  --star-delay: 0s;
  --star-duration: 7s;
  --star-opacity: 0.7;
  position: absolute;
  z-index: 4;
  width: var(--star-size);
  height: var(--star-size);
  transform: translate(-50%, -50%);
  opacity: var(--star-opacity);
  animation: star-breathe var(--star-duration) ease-in-out var(--star-delay) infinite;
}

.star::before,
.star::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}

.star-pin {
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 6px color-mix(in srgb, currentColor 70%, transparent);
}

.star-soft {
  border-radius: 50%;
  background: radial-gradient(circle, currentColor 0 18%, color-mix(in srgb, currentColor 42%, transparent) 32%, transparent 70%);
  filter: blur(0.2px);
}

.star-diamond {
  width: calc(var(--star-size) * 0.62);
  height: calc(var(--star-size) * 0.62);
  transform: translate(-50%, -50%) rotate(45deg);
  border-radius: 1px;
  background: currentColor;
  box-shadow:
    0 0 7px color-mix(in srgb, currentColor 68%, transparent),
    0 0 16px color-mix(in srgb, currentColor 34%, transparent);
}

.star-flare {
  border-radius: 50%;
  background: currentColor;
  box-shadow:
    0 0 8px color-mix(in srgb, currentColor 88%, transparent),
    0 0 22px color-mix(in srgb, currentColor 38%, transparent);
}

.star-flare::before {
  width: calc(var(--star-size) * 4.8);
  height: 1px;
  background: linear-gradient(90deg, transparent, color-mix(in srgb, currentColor 72%, transparent), currentColor, color-mix(in srgb, currentColor 72%, transparent), transparent);
  filter: blur(0.25px);
}

.star-flare::after {
  width: 1px;
  height: calc(var(--star-size) * 4.8);
  background: linear-gradient(180deg, transparent, color-mix(in srgb, currentColor 70%, transparent), currentColor, color-mix(in srgb, currentColor 70%, transparent), transparent);
  filter: blur(0.25px);
}

.star-ivory { color: #fffaf2; }
.star-cyan { color: #63fff3; }
.star-rose { color: #ff9fd0; }

.orbit {
  position: absolute;
  z-index: 4;
  left: clamp(34px, 7vw, 108px);
  top: clamp(70px, 11vh, 155px);
  width: clamp(56px, 6vw, 82px);
  aspect-ratio: 1;
  opacity: 0.9;
  animation: orbit-float 9s ease-in-out infinite alternate;
}

.orbit-ring,
.orbit-core {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border-radius: 50%;
}

.orbit-ring {
  width: 100%;
  height: 100%;
  border: 1.5px solid rgba(255, 253, 248, 0.86);
  box-shadow:
    0 0 10px rgba(255, 255, 255, 0.14),
    inset 0 0 10px rgba(255, 255, 255, 0.05);
}

.orbit-core {
  width: 29%;
  height: 29%;
  background: #fffaf2;
  box-shadow: 0 0 11px rgba(255, 250, 242, 0.34);
}

.cloud-bank {
  z-index: 2;
  top: auto;
  height: 34%;
  filter: blur(18px) saturate(108%);
  opacity: 0.35;
}

.cloud-bank-left {
  right: 43%;
  bottom: -19%;
  left: -9%;
  background:
    radial-gradient(ellipse at 18% 32%, rgba(224, 126, 174, 0.46), transparent 37%),
    radial-gradient(ellipse at 43% 18%, rgba(131, 112, 175, 0.42), transparent 39%),
    radial-gradient(ellipse at 68% 41%, rgba(73, 113, 143, 0.38), transparent 42%);
}

.cloud-bank-right {
  right: -8%;
  bottom: -18%;
  left: 48%;
  background:
    radial-gradient(ellipse at 24% 29%, rgba(83, 111, 166, 0.44), transparent 39%),
    radial-gradient(ellipse at 52% 22%, rgba(163, 107, 180, 0.46), transparent 40%),
    radial-gradient(ellipse at 78% 42%, rgba(239, 127, 184, 0.4), transparent 38%);
}

.cosmos-vignette {
  z-index: 7;
  background:
    radial-gradient(ellipse at center, transparent 44%, rgba(0, 0, 0, 0.18) 76%, rgba(0, 0, 0, 0.52) 100%),
    linear-gradient(180deg, rgba(0, 0, 0, 0.26), transparent 24%, transparent 76%, rgba(3, 6, 18, 0.35));
}

.cosmos-grain {
  z-index: 8;
  opacity: 0.017;
  mix-blend-mode: soft-light;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 180 180' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='.8'/%3E%3C/svg%3E");
}

@keyframes star-breathe {
  0%, 100% { opacity: calc(var(--star-opacity) * 0.58); transform: translate(-50%, -50%) scale(0.84); }
  48% { opacity: var(--star-opacity); transform: translate(-50%, -50%) scale(1.08); }
  65% { opacity: calc(var(--star-opacity) * 0.76); }
}

@keyframes dust-drift {
  from { transform: translate3d(0, 0, 0); }
  to { transform: translate3d(-72px, 38px, 0); }
}

@keyframes nebula-cyan-drift {
  from { transform: translate3d(-2%, -2%, 0) scale(0.97); }
  to { transform: translate3d(7%, 4%, 0) scale(1.05); }
}

@keyframes nebula-violet-drift {
  from { transform: translate3d(3%, 1%, 0) scale(0.98); }
  to { transform: translate3d(-6%, -3%, 0) scale(1.06); }
}

@keyframes nebula-rose-breathe {
  from { transform: scale(0.96); opacity: 0.24; }
  to { transform: scale(1.07); opacity: 0.36; }
}

@keyframes orbit-float {
  from { transform: translate3d(0, -3px, 0); }
  to { transform: translate3d(0, 8px, 0); }
}

@media (max-width: 720px) {
  .orbit {
    left: 30px;
    top: 70px;
    width: 58px;
  }

  .cosmos-nebula {
    filter: blur(58px) saturate(125%);
  }

  .nebula-cyan {
    inset: 34% 25% 11% -44%;
  }

  .nebula-violet {
    inset: 43% -45% 5% 32%;
  }

  .star:nth-of-type(n + 17) {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cosmos *,
  .cosmos::before,
  .cosmos::after {
    animation: none !important;
  }
}
</style>
