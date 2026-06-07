import type { PrimeVueConfiguration } from 'primevue/config'
import { loomiPreset } from './preset'

export const primeVueConfig: PrimeVueConfiguration = {
  theme: {
    preset: loomiPreset,
    options: {
      darkModeSelector: '.loomidbx-dark',
      cssLayer: {
        name: 'primevue',
        order: 'theme, base, primevue',
      },
    },
  },
}
