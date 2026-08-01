<template>
  <div class="cosmos" aria-hidden="true">
    <div class="cosmos-sky"></div>
    <div class="cosmos-haze"></div>

    <div class="orbit-mark">
      <span class="orbit-ring"></span>
      <span class="orbit-core"></span>
    </div>

    <span
      v-for="star in majorStars"
      :key="`major-${star.id}`"
      class="major-star"
      :class="`major-star--${star.color}`"
      :style="majorStarStyle(star)"
    ></span>

    <span
      v-for="star in smallStars"
      :key="`small-${star.id}`"
      class="small-star"
      :class="`small-star--${star.color}`"
      :style="smallStarStyle(star)"
    ></span>

    <div class="cosmos-vignette"></div>
  </div>
</template>

<script setup lang="ts">
type StarColor = 'ivory' | 'cyan' | 'rose'

type MajorStar = {
  id: number
  x: number
  y: number
  size: number
  color: StarColor
  opacity: number
  pulse?: boolean
}

type SmallStar = {
  id: number
  x: number
  y: number
  size: number
  color: StarColor
  opacity: number
}

const majorStars: MajorStar[] = [
  { id: 1, x: 5.8, y: 35.4, size: 28, color: 'cyan', opacity: 0.84 },
  { id: 2, x: 13.2, y: 47.8, size: 18, color: 'cyan', opacity: 0.76, pulse: true },
  { id: 3, x: 18.8, y: 41.6, size: 13, color: 'cyan', opacity: 0.72 },
  { id: 4, x: 46.5, y: 28.2, size: 16, color: 'rose', opacity: 0.62 },
  { id: 5, x: 50.8, y: 34.4, size: 30, color: 'cyan', opacity: 0.93, pulse: true },
  { id: 6, x: 60.2, y: 23.8, size: 32, color: 'ivory', opacity: 0.9 },
  { id: 7, x: 87.6, y: 26.2, size: 36, color: 'cyan', opacity: 0.95, pulse: true },
  { id: 8, x: 80.7, y: 45.6, size: 22, color: 'cyan', opacity: 0.82 },
  { id: 9, x: 2.6, y: 53.4, size: 34, color: 'rose', opacity: 0.82 },
  { id: 10, x: 94.4, y: 51.2, size: 27, color: 'ivory', opacity: 0.78 },
  { id: 11, x: 17.5, y: 69.4, size: 25, color: 'cyan', opacity: 0.72 },
  { id: 12, x: 82.8, y: 73.6, size: 25, color: 'ivory', opacity: 0.74 }
]

const smallStars: SmallStar[] = [
  { id: 1, x: 9, y: 29, size: 1.1, color: 'ivory', opacity: 0.52 },
  { id: 2, x: 15.5, y: 38.5, size: 1.4, color: 'cyan', opacity: 0.66 },
  { id: 3, x: 23.2, y: 31.6, size: 1.1, color: 'rose', opacity: 0.48 },
  { id: 4, x: 31, y: 43.2, size: 1.5, color: 'ivory', opacity: 0.66 },
  { id: 5, x: 39.6, y: 25.6, size: 1, color: 'ivory', opacity: 0.5 },
  { id: 6, x: 48.4, y: 39.2, size: 1.2, color: 'cyan', opacity: 0.58 },
  { id: 7, x: 56.8, y: 31.4, size: 1, color: 'ivory', opacity: 0.48 },
  { id: 8, x: 65.2, y: 43.8, size: 1.4, color: 'rose', opacity: 0.5 },
  { id: 9, x: 72.4, y: 29.6, size: 1.2, color: 'cyan', opacity: 0.57 },
  { id: 10, x: 79.1, y: 37.1, size: 1.1, color: 'ivory', opacity: 0.58 },
  { id: 11, x: 91.4, y: 33.4, size: 1.4, color: 'rose', opacity: 0.5 },
  { id: 12, x: 97.2, y: 43.6, size: 1.1, color: 'ivory', opacity: 0.56 },
  { id: 13, x: 7.2, y: 61.2, size: 1.2, color: 'ivory', opacity: 0.52 },
  { id: 14, x: 24.8, y: 58.6, size: 1.1, color: 'rose', opacity: 0.46 },
  { id: 15, x: 35.5, y: 66.8, size: 1.5, color: 'cyan', opacity: 0.62 },
  { id: 16, x: 46.4, y: 55.3, size: 1.1, color: 'ivory', opacity: 0.46 },
  { id: 17, x: 57.2, y: 63.6, size: 1.3, color: 'rose', opacity: 0.52 },
  { id: 18, x: 68.6, y: 57.2, size: 1.1, color: 'ivory', opacity: 0.48 },
  { id: 19, x: 75.8, y: 68.2, size: 1.5, color: 'cyan', opacity: 0.61 },
  { id: 20, x: 90.2, y: 62.8, size: 1.2, color: 'ivory', opacity: 0.5 },
  { id: 21, x: 11.6, y: 79.1, size: 1, color: 'rose', opacity: 0.42 },
  { id: 22, x: 28.4, y: 75.8, size: 1.2, color: 'ivory', opacity: 0.46 },
  { id: 23, x: 43.8, y: 82.4, size: 1.1, color: 'cyan', opacity: 0.5 },
  { id: 24, x: 61.4, y: 77.6, size: 1.3, color: 'ivory', opacity: 0.49 },
  { id: 25, x: 72.8, y: 86.3, size: 1, color: 'rose', opacity: 0.4 },
  { id: 26, x: 93.1, y: 81.2, size: 1.3, color: 'cyan', opacity: 0.54 }
]

function majorStarStyle(star: MajorStar) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    '--star-size': `${star.size}px`,
    '--star-opacity': star.opacity,
    '--star-animation': star.pulse ? 'star-breathe 9s ease-in-out infinite' : 'none'
  }
}

function smallStarStyle(star: SmallStar) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    width: `${star.size}px`,
    height: `${star.size}px`,
    opacity: star.opacity
  }
}
</script>

<style scoped>
.cosmos {
  position: absolute;
  inset: 0;
  z-index: 0;
  min-height: 100vh;
  height: 100%;
  overflow: hidden;
  pointer-events: none;
  contain: paint;
  background: #010305;
}

.cosmos-sky,
.cosmos-haze,
.cosmos-vignette {
  position: absolute;
  inset: 0;
}

.cosmos-sky {
  background:
    radial-gradient(ellipse at 48% 44%, rgba(8, 78, 83, 0.42) 0%, rgba(4, 38, 43, 0.24) 30%, transparent 58%),
    radial-gradient(ellipse at 20% 62%, rgba(29, 113, 115, 0.24) 0%, transparent 46%),
    radial-gradient(ellipse at 82% 65%, rgba(95, 43, 103, 0.24) 0%, transparent 48%),
    linear-gradient(
      180deg,
      #010204 0%,
      #010407 20%,
      #021014 38%,
      #07343a 52%,
      #76506f 70%,
      #332848 88%,
      #121522 100%
    );
}

.cosmos-haze {
  opacity: 0.88;
  background:
    radial-gradient(ellipse at 50% 54%, rgba(93, 218, 210, 0.14) 0%, transparent 38%),
    radial-gradient(ellipse at 50% 70%, rgba(221, 111, 169, 0.16) 0%, transparent 44%),
    linear-gradient(180deg, transparent 0 43%, rgba(0, 14, 18, 0.1) 52%, transparent 75%);
}

.orbit-mark {
  position: absolute;
  left: clamp(52px, 8vw, 118px);
  top: clamp(88px, 12vh, 168px);
  width: 76px;
  height: 76px;
  opacity: 0.9;
}

.orbit-ring,
.orbit-core {
  position: absolute;
  left: 50%;
  top: 50%;
  border-radius: 50%;
  transform: translate(-50%, -50%);
}

.orbit-ring {
  width: 70px;
  height: 70px;
  border: 2px solid rgba(244, 242, 235, 0.86);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.08),
    0 0 20px rgba(214, 240, 238, 0.06);
}

.orbit-core {
  width: 20px;
  height: 20px;
  background: #f4f1e8;
  box-shadow: 0 0 12px rgba(255, 255, 255, 0.2);
}

.major-star {
  --star-color: 235, 242, 238;
  position: absolute;
  width: 3px;
  height: 3px;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  opacity: var(--star-opacity);
  background: rgb(var(--star-color));
  box-shadow:
    0 0 5px 1px rgba(var(--star-color), 0.55),
    0 0 16px 3px rgba(var(--star-color), 0.18);
  animation: var(--star-animation);
}

.major-star::before,
.major-star::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border-radius: 999px;
}

.major-star::before {
  width: var(--star-size);
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(var(--star-color), 0.08) 20%,
    rgba(var(--star-color), 0.72) 46%,
    rgba(var(--star-color), 1) 50%,
    rgba(var(--star-color), 0.72) 54%,
    rgba(var(--star-color), 0.08) 80%,
    transparent 100%
  );
}

.major-star::after {
  width: 1px;
  height: calc(var(--star-size) * 1.15);
  background: linear-gradient(
    180deg,
    transparent 0%,
    rgba(var(--star-color), 0.06) 18%,
    rgba(var(--star-color), 0.74) 46%,
    rgba(var(--star-color), 1) 50%,
    rgba(var(--star-color), 0.74) 54%,
    rgba(var(--star-color), 0.06) 82%,
    transparent 100%
  );
}

.major-star--ivory {
  --star-color: 244, 242, 232;
}

.major-star--cyan {
  --star-color: 90, 246, 239;
}

.major-star--rose {
  --star-color: 255, 161, 199;
}

.small-star {
  position: absolute;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  background: currentColor;
  box-shadow: 0 0 5px currentColor;
}

.small-star--ivory {
  color: rgba(244, 242, 232, 0.92);
}

.small-star--cyan {
  color: rgba(102, 244, 237, 0.92);
}

.small-star--rose {
  color: rgba(255, 162, 201, 0.88);
}

.cosmos-vignette {
  box-shadow:
    inset 0 0 180px rgba(0, 0, 0, 0.42),
    inset 0 -160px 220px rgba(4, 4, 13, 0.34);
}

@keyframes star-breathe {
  0%,
  100% {
    opacity: calc(var(--star-opacity) * 0.72);
  }
  50% {
    opacity: var(--star-opacity);
  }
}

@media (max-width: 720px) {
  .orbit-mark {
    left: 44px;
    top: 96px;
    transform: scale(0.82);
    transform-origin: top left;
  }

  .major-star:nth-of-type(n + 9),
  .small-star:nth-of-type(n + 17) {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .major-star {
    animation: none !important;
  }
}
</style>
