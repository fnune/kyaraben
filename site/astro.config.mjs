import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'

export default defineConfig({
  integrations: [
    starlight({
      title: 'Kyaraben',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/fnune/kyaraben',
        },
      ],
      sidebar: [
        { label: 'Getting started', slug: 'getting-started' },
        { label: 'Using kyaraben', slug: 'using-kyaraben' },
        { label: 'Configuration', slug: 'configuration' },
        { label: 'Save sync', slug: 'sync' },
        { label: 'How it works', slug: 'how-it-works' },
        { label: 'Updating', slug: 'updating' },
        {
          label: 'Contributing',
          items: [
            { label: 'Development guide', slug: 'contributing' },
            {
              label: 'Emulator config reference',
              slug: 'contributing/emulator-config-reference',
            },
          ],
        },
      ],
    }),
  ],
})
