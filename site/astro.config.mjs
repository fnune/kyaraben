import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  integrations: [
    starlight({
      title: "Kyaraben",
      favicon: "/kyaraben-app-logo.svg",
      components: {
        SiteTitle: "./src/components/SiteTitle.astro",
      },
      customCss: ["./src/styles/custom.css"],
      expressiveCode: {
        themes: ["dark-plus"],
      },
      head: [
        {
          tag: "link",
          attrs: {
            rel: "preconnect",
            href: "https://fonts.googleapis.com",
          },
        },
        {
          tag: "link",
          attrs: {
            rel: "preconnect",
            href: "https://fonts.gstatic.com",
            crossorigin: true,
          },
        },
        {
          tag: "link",
          attrs: {
            rel: "stylesheet",
            href: "https://fonts.googleapis.com/css2?family=DM+Serif+Display:ital@0;1&family=IBM+Plex+Mono&family=IBM+Plex+Sans:ital,wght@0,400;0,500;0,600;1,400&display=swap",
          },
        },
      ],
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/fnune/kyaraben",
        },
      ],
      sidebar: [
        { label: "Getting started", slug: "getting-started" },
        { label: "App reference", slug: "using-the-app" },
        { label: "CLI reference", slug: "using-the-cli" },
        { label: "Sync", slug: "sync" },
        { label: "Contributing", slug: "contributing" },
      ],
    }),
  ],
});
