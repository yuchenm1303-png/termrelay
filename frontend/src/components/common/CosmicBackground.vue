<template>
  <div ref="cosmosRef" class="cosmos" aria-hidden="true">
    <div class="cosmos-sky"></div>
    <div class="cosmos-nebula cosmos-nebula-cyan"></div>
    <div class="cosmos-nebula cosmos-nebula-violet"></div>

    <div class="parallax-layer parallax-layer-far">
      <span
        v-for="star in smallStars"
        :key="`small-${star.id}`"
        class="small-star"
        :class="`small-star--${star.color}`"
        :style="smallStarStyle(star)"
      ></span>
    </div>

    <div class="parallax-layer parallax-layer-near">
      <span
        v-for="star in majorStars"
        :key="`major-${star.id}`"
        class="major-star"
        :class="`major-star--${star.color}`"
        :style="majorStarStyle(star)"
      ></span>

      <div class="upper-right-stars">
        <span class="hero-flare hero-flare-main"></span>
        <span class="hero-flare hero-flare-secondary"></span>
        <span class="hero-dot hero-dot-a"></span>
        <span class="hero-dot hero-dot-b"></span>
        <span class="hero-dot hero-dot-c"></span>
        <span class="hero-trail"></span>
      </div>
    </div>

    <div class="cosmos-vignette"></div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

type StarColor = 'ivory' | 'cyan' | 'rose'

type Star = {
  id: number
  x: number
  y: number
  size: number
  color: StarColor
  opacity: number
}

const cosmosRef = ref<HTMLElement | null>(null)
let animationFrame = 0
let reduceMotionQuery: MediaQueryList | null = null

const majorStars: Star[] = [
  { id: 1, x: 8, y: 42, size: 18, color: 'cyan', opacity: 0.42 },
  { id: 2, x: 54, y: 31, size: 17, color: 'ivory', opacity: 0.48 },
  { id: 3, x: 88, y: 50, size: 24, color: 'cyan', opacity: 0.56 },
  { id: 4, x: 68, y: 69, size: 16, color: 'rose', opacity: 0.34 },
  { id: 5, x: 17, y: 78, size: 16, color: 'cyan', opacity: 0.32 }
]

const smallStars: Star[] = [
  { id: 1, x: 11, y: 24, size: 0.8, color: 'ivory', opacity: 0.28 },
  { id: 2, x: 21, y: 34, size: 0.8, color: 'cyan', opacity: 0.31 },
  { id: 3, x: 37, y: 22, size: 0.7, color: 'ivory', opacity: 0.24 },
  { id: 4, x: 47, y: 43, size: 0.8, color: 'rose', opacity: 0.26 },
  { id: 5, x: 63, y: 26, size: 0.7, color: 'ivory', opacity: 0.25 },
  { id: 6, x: 76, y: 36, size: 0.8, color: 'cyan', opacity: 0.29 },
  { id: 7, x: 95, y: 33, size: 0.8, color: 'ivory', opacity: 0.27 },
  { id: 8, x: 31, y: 57, size: 0.7, color: 'ivory', opacity: 0.22 },
  { id: 9, x: 58, y: 55, size: 0.8, color: 'cyan', opacity: 0.27 },
  { id: 10, x: 82, y: 63, size: 0.8, color: 'rose', opacity: 0.23 },
  { id: 11, x: 12, y: 67, size: 0.7, color: 'cyan', opacity: 0.23 },
  { id: 12, x: 42, y: 74, size: 0.8, color: 'ivory', opacity: 0.22 },
  { id: 13, x: 72, y: 81, size: 0.8, color: 'ivory', opacity: 0.24 },
  { id: 14, x: 94, y: 86, size: 0.7, color: 'cyan', opacity: 0.25 }
]

function majorStarStyle(star: Star) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    '--star-size': `${star.size}px`,
    '--star-opacity': star.opacity
  }
}

function smallStarStyle(star: Star) {
  return {
    left: `${star.x}%`,
    top: `${star.y}%`,
    width: `${star.size}px`,
    height: `${star.size}px`,
    opacity: star.opacity
  }
}

function renderParallax() {
  animationFrame = 0
  if (!cosmosRef.value) return

  const reduceMotion = reduceMotionQuery?.matches ?? false
  const maxScroll = Math.max(
    document.documentElement.scrollHeight - document.documentElement.clientHeight,
    0
  )
  const rawScroll = window.scrollY || document.documentElement.scrollTop || 0
  const scrollTop = Math.min(Math.max(rawScroll, 0), maxScroll)

  const farOffset = reduceMotion ? 0 : -Math.min(scrollTop * 0.035, 96)
  const nearOffset = reduceMotion ? 0 : -Math.min(scrollTop * 0.075, 210)

  cosmosRef.value.style.setProperty('--far-offset', `${farOffset.toFixed(2)}px`)
  cosmosRef.value.style.setProperty('--near-offset', `${nearOffset.toFixed(2)}px`)
}

function requestParallaxUpdate() {
  if (animationFrame) return
  animationFrame = window.requestAnimationFrame(renderParallax)
}

onMounted(() => {
  reduceMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  renderParallax()
  window.addEventListener('scroll', requestParallaxUpdate, { passive: true })
  window.addEventListener('resize', requestParallaxUpdate, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', requestParallaxUpdate)
  window.removeEventListener('resize', requestParallaxUpdate)
  if (animationFrame) window.cancelAnimationFrame(animationFrame)
})
</script>

<style scoped>
.cosmos {
  --far-offset: 0px;
  --near-offset: 0px;
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  background: #010304;
  transform: translateZ(0);
}

.cosmos-sky,
.cosmos-nebula,
.cosmos-vignette {
  position: absolute;
  inset: 0;
}

.cosmos-sky {
  background:
    linear-gradient(
      180deg,
      #010203 0%,
      #010405 30%,
      #031116 51%,
      #0a3038 69%,
      #3b344c 86%,
      #161722 100%
    );
}

.cosmos-nebula {
  opacity: 0.55;
}

.cosmos-nebula-cyan {
  background: radial-gradient(
    ellipse at 74% 48%,
    rgba(61, 189, 187, 0.1) 0%,
    rgba(25, 89, 96, 0.06) 34%,
    transparent 63%
  );
}

.cosmos-nebula-violet {
  background: radial-gradient(
    ellipse at 26% 88%,
    rgba(150, 84, 147, 0.11) 0%,
    rgba(68, 50, 105, 0.05) 38%,
    transparent 66%
  );
}

.parallax-layer {
  position: absolute;
  right: 0;
  left: 0;
  will-change: transform;
  backface-visibility: hidden;
}

.parallax-layer-far {
  top: -120px;
  bottom: -120px;
  transform: translate3d(0, var(--far-offset), 0);
}

.parallax-layer-near {
  top: -240px;
  bottom: -240px;
  transform: translate3d(0, var(--near-offset), 0);
}

.major-star,
.hero-flare {
  --star-color: 241, 243, 238;
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: rgb(var(--star-color));
  box-shadow:
    0 0 4px rgba(var(--star-color), 0.48),
    0 0 12px rgba(var(--star-color), 0.12);
}

.major-star {
  transform: translate(-50%, -50%);
  opacity: var(--star-opacity);
}

.major-star::before,
.major-star::after,
.hero-flare::before,
.hero-flare::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border-radius: 999px;
}

.major-star::before,
.hero-flare::before {
  width: var(--star-size);
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(var(--star-color), 0.08) 22%,
    rgba(var(--star-color), 0.7) 47%,
    rgba(var(--star-color), 1) 50%,
    rgba(var(--star-color), 0.7) 53%,
    rgba(var(--star-color), 0.08) 78%,
    transparent
  );
}

.major-star::after,
.hero-flare::after {
  width: 1px;
  height: calc(var(--star-size) * 1.14);
  background: linear-gradient(
    180deg,
    transparent,
    rgba(var(--star-color), 0.06) 20%,
    rgba(var(--star-color), 0.68) 47%,
    rgba(var(--star-color), 1) 50%,
    rgba(var(--star-color), 0.68) 53%,
    rgba(var(--star-color), 0.06) 80%,
    transparent
  );
}

.major-star--ivory {
  --star-color: 244, 242, 232;
}

.major-star--cyan {
  --star-color: 93, 237, 232;
}

.major-star--rose {
  --star-color: 244, 161, 193;
}

.small-star {
  position: absolute;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  background: currentColor;
  box-shadow: 0 0 4px currentColor;
}

.small-star--ivory {
  color: rgba(244, 242, 232, 0.9);
}

.small-star--cyan {
  color: rgba(104, 235, 229, 0.9);
}

.small-star--rose {
  color: rgba(244, 165, 196, 0.86);
}

.upper-right-stars {
  position: absolute;
  right: clamp(34px, 7vw, 112px);
  top: clamp(92px, 12vh, 150px);
  width: clamp(190px, 24vw, 310px);
  height: clamp(128px, 18vw, 214px);
  opacity: 0.84;
}

.hero-flare-main {
  --star-color: 92, 240, 234;
  --star-size: 48px;
  right: 8%;
  top: 22%;
}

.hero-flare-secondary {
  --star-color: 244, 242, 232;
  --star-size: 25px;
  left: 28%;
  bottom: 17%;
  opacity: 0.68;
}

.hero-dot {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: rgba(236, 245, 241, 0.7);
  box-shadow: 0 0 7px rgba(219, 245, 241, 0.2);
}

.hero-dot-a {
  left: 8%;
  top: 19%;
}

.hero-dot-b {
  left: 52%;
  top: 8%;
  background: rgba(246, 174, 203, 0.68);
}

.hero-dot-c {
  right: 24%;
  bottom: 7%;
  background: rgba(97, 235, 230, 0.72);
}

.hero-trail {
  position: absolute;
  left: 17%;
  top: 52%;
  width: 57%;
  height: 1px;
  opacity: 0.14;
  transform: rotate(-18deg);
  transform-origin: left center;
  background: linear-gradient(90deg, transparent, rgba(113, 224, 218, 0.5), transparent);
}

.cosmos-vignette {
  box-shadow:
    inset 0 0 190px rgba(0, 0, 0, 0.5),
    inset 0 -140px 220px rgba(5, 5, 15, 0.34);
}

@media (max-width: 760px) {
  .parallax-layer-far {
    top: -100px;
    bottom: -100px;
  }

  .parallax-layer-near {
    top: -220px;
    bottom: -220px;
  }

  .upper-right-stars {
    right: -18px;
    top: 92px;
    transform: scale(0.86);
    transform-origin: top right;
    opacity: 0.58;
  }

  .major-star:nth-of-type(n + 4),
  .small-star:nth-of-type(n + 10) {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .parallax-layer-far,
  .parallax-layer-near {
    transform: none;
  }
}
</style>
