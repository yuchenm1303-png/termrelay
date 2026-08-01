<template>
  <div class="cosmos" aria-hidden="true">
    <div class="cosmos-sky"></div>
    <div class="cosmos-color-wash"></div>
    <div class="cosmos-content-veil"></div>

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
}

type SmallStar = {
  id: number
  x: number
  y: number
  size: number
  color: StarColor
  opacity: number
}

/*
 * Keep the bright stars sparse and mostly outside the reading column.
 * The reference uses a few deliberate four-point flares rather than a particle field.
 */
const majorStars: MajorStar[] = [
  { id: 1, x: 60, y: 24, size: 28, color: 'ivory', opacity: 0.72 },
  { id: 2, x: 87, y: 27, size: 34, color: 'cyan', opacity: 0.82 },
  { id: 3, x: 75, y: 39, size: 20, color: 'rose', opacity: 0.48 },
  { id: 4, x: 91, y: 48, size: 26, color: 'cyan', opacity: 0.7 },
  { id: 5, x: 62, y: 61, size: 18, color: 'cyan', opacity: 0.58 },
  { id: 6, x: 84, y: 69, size: 24, color: 'ivory', opacity: 0.62 },
  { id: 7, x: 16, y: 76, size: 18, color: 'rose', opacity: 0.42 },
  { id: 8, x: 55, y: 86, size: 20, color: 'cyan', opacity: 0.48 }
]

const smallStars: SmallStar[] = [
  { id: 1, x: 10, y: 25, size: 0.8, color: 'ivory', opacity: 0.32 },
  { id: 2, x: 21, y: 31, size: 0.9, color: 'cyan', opacity: 0.4 },
  { id: 3, x: 39, y: 22, size: 0.8, color: 'ivory', opacity: 0.28 },
  { id: 4, x: 52, y: 34, size: 0.9, color: 'rose', opacity: 0.34 },
  { id: 5, x: 69, y: 30, size: 0.8, color: 'ivory', opacity: 0.3 },
  { id: 6, x: 79, y: 21, size: 0.8, color: 'cyan', opacity: 0.38 },
  { id: 7, x: 95, y: 36, size: 0.9, color: 'ivory', opacity: 0.34 },
  { id: 8, x: 33, y: 47, size: 0.8, color: 'ivory', opacity: 0.28 },
  { id: 9, x: 57, y: 51, size: 0.8, color: 'cyan', opacity: 0.34 },
  { id: 10, x: 82, y: 55, size: 0.9, color: 'rose', opacity: 0.3 },
  { id: 11, x: 12, y: 63, size: 0.8, color: 'cyan', opacity: 0.3 },
  { id: 12, x: 29, y: 68, size: 0.9, color: 'ivory', opacity: 0.28 },
  { id: 13, x: 47, y: 72, size: 0.8, color: 'rose', opacity: 0.26 },
  { id: 14, x: 72, y: 77, size: 0.9, color: 'ivory', opacity: 0.3 },
  { id: 15, x: 94, y: 83, size: 0.8, color: 'cyan', opacity: 0.32 }
]

function majorStarStyle(star: MajorStar) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    '--star-size': `${star.size}px`,
    '--star-opacity': star.opacity
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
  min-height: 100%;
  overflow: hidden;
  pointer-events: none;
  contain: paint;
  background: #010304;
}

.cosmos-sky,
.cosmos-color-wash,
.cosmos-content-veil,
.cosmos-vignette {
  position: absolute;
  inset: 0;
}

/* Nearly black above the fold; color only enters gradually lower down. */
.cosmos-sky {
  background:
    linear-gradient(
      180deg,
      #010203 0%,
      #010304 22%,
      #021014 42%,
      #073139 59%,
      #594157 75%,
      #27243a 91%,
      #11131d 100%
    );
}

.cosmos-color-wash {
  opacity: 0.7;
  background:
    radial-gradient(ellipse at 76% 42%, rgba(56, 196, 190, 0.09) 0%, transparent 31%),
    radial-gradient(ellipse at 18% 58%, rgba(23, 119, 121, 0.1) 0%, transparent 36%),
    radial-gradient(ellipse at 74% 77%, rgba(166, 78, 145, 0.11) 0%, transparent 39%);
}

/* A dark reading lane keeps the terminal typography and the sky in the same visual system. */
.cosmos-content-veil {
  background:
    linear-gradient(
      90deg,
      rgba(1, 3, 5, 0.74) 0%,
      rgba(1, 3, 5, 0.57) 42%,
      rgba(1, 3, 5, 0.24) 67%,
      rgba(1, 3, 5, 0.05) 100%
    ),
    linear-gradient(
      180deg,
      rgba(1, 3, 5, 0.28) 0%,
      transparent 28%,
      transparent 70%,
      rgba(5, 5, 14, 0.2) 100%
    );
}

/* Move the reference-inspired orbit into the negative space, away from the headline. */
.orbit-mark {
  position: absolute;
  right: clamp(58px, 10vw, 150px);
  top: clamp(105px, 13vh, 175px);
  width: 68px;
  height: 68px;
  opacity: 0.72;
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
  width: 64px;
  height: 64px;
  border: 1.5px solid rgba(244, 241, 233, 0.78);
  box-shadow: 0 0 16px rgba(204, 238, 235, 0.04);
}

.orbit-core {
  width: 17px;
  height: 17px;
  background: rgba(244, 241, 233, 0.92);
  box-shadow: 0 0 9px rgba(255, 255, 255, 0.16);
}

.major-star {
  --star-color: 240, 243, 238;
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  opacity: var(--star-opacity);
  background: rgb(var(--star-color));
  box-shadow:
    0 0 4px rgba(var(--star-color), 0.48),
    0 0 11px rgba(var(--star-color), 0.13);
}

.major-star::before,
.major-star::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}

.major-star::before {
  width: var(--star-size);
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(var(--star-color), 0.05) 24%,
    rgba(var(--star-color), 0.62) 47%,
    rgba(var(--star-color), 0.96) 50%,
    rgba(var(--star-color), 0.62) 53%,
    rgba(var(--star-color), 0.05) 76%,
    transparent
  );
}

.major-star::after {
  width: 1px;
  height: calc(var(--star-size) * 1.18);
  background: linear-gradient(
    180deg,
    transparent,
    rgba(var(--star-color), 0.04) 22%,
    rgba(var(--star-color), 0.58) 47%,
    rgba(var(--star-color), 0.96) 50%,
    rgba(var(--star-color), 0.58) 53%,
    rgba(var(--star-color), 0.04) 78%,
    transparent
  );
}

.major-star--ivory {
  --star-color: 240, 239, 229;
}

.major-star--cyan {
  --star-color: 86, 230, 224;
}

.major-star--rose {
  --star-color: 235, 145, 181;
}

.small-star {
  position: absolute;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  background: currentColor;
  box-shadow: 0 0 3px currentColor;
}

.small-star--ivory {
  color: rgba(240, 239, 229, 0.9);
}

.small-star--cyan {
  color: rgba(94, 224, 218, 0.88);
}

.small-star--rose {
  color: rgba(235, 145, 181, 0.82);
}

.cosmos-vignette {
  box-shadow:
    inset 0 0 170px rgba(0, 0, 0, 0.46),
    inset 0 -170px 230px rgba(3, 4, 12, 0.3);
}

@media (max-width: 940px) {
  .cosmos-content-veil {
    background:
      linear-gradient(
        90deg,
        rgba(1, 3, 5, 0.72) 0%,
        rgba(1, 3, 5, 0.47) 68%,
        rgba(1, 3, 5, 0.14) 100%
      );
  }

  .orbit-mark {
    right: 56px;
    top: 128px;
    transform: scale(0.86);
    transform-origin: top right;
  }
}

@media (max-width: 720px) {
  .orbit-mark {
    display: none;
  }

  .major-star:nth-of-type(n + 6),
  .small-star:nth-of-type(n + 11) {
    display: none;
  }

  .cosmos-color-wash {
    opacity: 0.5;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cosmos * {
    animation: none !important;
  }
}
</style>
