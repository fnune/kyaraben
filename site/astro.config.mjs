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
        { label: 'App reference', slug: 'using-the-app' },
        { label: 'CLI reference', slug: 'using-the-cli' },
        { label: 'Save sync', slug: 'sync' },
        { label: 'Contributing', slug: 'contributing' },
      ],
    }),
  ],
})
