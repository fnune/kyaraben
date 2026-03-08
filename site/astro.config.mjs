import { defineConfig, passthroughImageService } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  image: {
    service: passthroughImageService(),
  },
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
        { label: "Download", slug: "download" },
        {
          label: "Guides",
          items: [
            { label: "Desktop and Steam Deck", slug: "guides/desktop-and-steam-deck" },
            { label: "Headless server", slug: "guides/headless-server" },
            { label: "Adding a NextUI device", slug: "guides/nextui-device" },
            { label: "Migrating from EmuDeck", slug: "guides/migrating-from-emudeck" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "App", slug: "using-the-app" },
            { label: "CLI", slug: "using-the-cli" },
            { label: "Collection", slug: "collection" },
          ],
        },
        {
          label: "Integrations",
          items: [{ label: "NextUI", slug: "nextui" }],
        },
        {
          label: "Project",
          items: [
            { label: "Contributing", slug: "contributing" },
            { label: "Support", slug: "support" },
          ],
        },
      ],
    }),
  ],
});
