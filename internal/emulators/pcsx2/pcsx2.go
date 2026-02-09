// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package pcsx2

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDPCSX2,
		Name:    "PCSX2",
		Systems: []model.SystemID{model.SystemIDPS2},
		Package: model.AppImageRef("pcsx2"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one BIOS required",
			Provisions:  ps2BIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "pcsx2",
			GenericName: "PlayStation 2 Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " -fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.StandardPathUsage(),
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "PCSX2/inis/PCSX2.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"UI", "SettingsVersion"}, Value: "1"},
			{Path: []string{"UI", "SetupWizardIncomplete"}, Value: "false"},
			{Path: []string{"Folders", "Bios"}, Value: store.SystemBiosDir(model.SystemIDPS2)},
			{Path: []string{"Folders", "MemoryCards"}, Value: store.SystemSavesDir(model.SystemIDPS2)},
			{Path: []string{"Folders", "Savestates"}, Value: store.EmulatorStatesDir(model.EmulatorIDPCSX2)},
			{Path: []string{"Folders", "Snapshots"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDPCSX2)},
			{Path: []string{"GameList", "RecursivePaths"}, Value: store.SystemRomsDir(model.SystemIDPS2)},
		},
	}}, nil
}

var ps2BIOSProvisions = []model.Provision{
	{Kind: model.ProvisionBIOS, Filename: "ps2-0100j-20000117.bin", Description: "Japan v1.00", Hashes: []string{"acf4730ceb38ac9d8c7d8e21f2614600"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0100jd-20000117.bin", Description: "Japan v1.00 debug", Hashes: []string{"32f2e4d5ff5ee11072a6bc45530f5765"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0101j-20000217.bin", Description: "Japan v1.01", Hashes: []string{"b1459d7446c69e3e97e6ace3ae23dd1c"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0101jd-20000217.bin", Description: "Japan v1.01 debug", Hashes: []string{"acf9968c8f596d2b15f42272082513d1"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0101xd-20000224.bin", Description: "Export/TOOL debug", Hashes: []string{"d3f1853a16c2ec18f3cd1ae655213308"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0110a-20000727.bin", Description: "Asia v1.10", Hashes: []string{"a20c97c02210f16678ca3010127caf36"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0110ad-20000727.bin", Description: "Asia v1.10 debug", Hashes: []string{"63e6fd9b3c72e0d7b920e80cf76645cd"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0120a-20000902.bin", Description: "Asia v1.20", Hashes: []string{"8db2fbbac7413bf3e7154c1e0715e565"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0120e-20000902.bin", Description: "Europe v1.20", Hashes: []string{"b7fa11e87d51752a98b38e3e691cbf17"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0120ed-20000902.bin", Description: "Europe v1.20 debug", Hashes: []string{"91c87cb2f2eb6ce529a2360f80ce2457"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0120ed-20000902-20030110.bin", Description: "Europe v1.20 debug", Hashes: []string{"3016b3dd42148a67e2c048595ca4d7ce"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0120j-20001027-185015.bin", Description: "Japan v1.20", Hashes: []string{"f63bc530bd7ad7c026fcd6f7bd0d9525"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0120j-20001027-191435.bin", Description: "Japan v1.20", Hashes: []string{"cee06bd68c333fc5768244eae77e4495"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0150a-20001228.bin", Description: "Asia v1.50", Hashes: []string{"8accc3c49ac45f5ae2c5db0adc854633"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0150ad-20001228-20030520.bin", Description: "Asia v1.50 debug", Hashes: []string{"0bf988e9c7aaa4c051805b0fa6eb3387"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0150e-20001228.bin", Description: "Europe v1.50", Hashes: []string{"838544f12de9b0abc90811279ee223c8"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0150ed-20001228-20030520.bin", Description: "Europe v1.50 debug", Hashes: []string{"6f9a6feb749f0533aaae2cc45090b0ed"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0150j-20010118.bin", Description: "Japan v1.50", Hashes: []string{"815ac991d8bc3b364696bead3457de7d"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0150jd-20010118.bin", Description: "Japan v1.50 debug", Hashes: []string{"bb6bbc850458fff08af30e969ffd0175"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160a-20010427.bin", Description: "Asia v1.60", Hashes: []string{"b107b5710042abe887c0f6175f6e94bb"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160e-20010704.bin", Description: "Europe v1.60", Hashes: []string{"491209dd815ceee9de02dbbc408c06d6"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160e-20011004.bin", Description: "Europe v1.60", Hashes: []string{"8359638e857c8bc18c3c18ac17d9cc3c"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160h-20010730.bin", Description: "China/HK v1.60", Hashes: []string{"352d2ff9b3f68be7e6fa7e6dd8389346"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160h-20020426.bin", Description: "China/HK v1.60", Hashes: []string{"315a4003535dfda689752cb25f24785c"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160j-20010427.bin", Description: "Japan v1.60", Hashes: []string{"ab55cceea548303c22c72570cfd4dd71"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0160j-20020426.bin", Description: "Japan v1.60", Hashes: []string{"72da56fccb8fcd77bba16d1b6f479914"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0170ad-20030325.bin", Description: "Asia v1.70 debug", Hashes: []string{"eb960de68f0c0f7f9fa083e9f79d0360"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0170e-20030227.bin", Description: "Europe v1.70", Hashes: []string{"666018ffec65c5c7e04796081295c6c7"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0170ed-20030227.bin", Description: "Europe v1.70 debug", Hashes: []string{"666018ffec65c5c7e04796081295c6c7"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0170j-20030206.bin", Description: "Japan v1.70", Hashes: []string{"312ad4816c232a9606e56f946bc0678a"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0180cd-20030224.bin", Description: "Korea/China debug", Hashes: []string{"240d4c5ddd4b54069bdc4a3cd2faf99d"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0180j-20031028.bin", Description: "Japan v1.80", Hashes: []string{"1c6cd089e6c83da618fbf2a081eb4888"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190a-20031128.bin", Description: "Asia v1.90", Hashes: []string{"4587a1dd58d6805265bfc68f0fef2732"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190ad-20031128.bin", Description: "Asia v1.90 debug", Hashes: []string{"30ab5f78d80431a862283642b71f8fb3"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190e-20031031.bin", Description: "Europe v1.90", Hashes: []string{"d1c4ac413fdf57cd93585ff19cf50df0"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190ed-20031031.bin", Description: "Europe v1.90 debug", Hashes: []string{"0e440a76cd950cbbb1d576242e0429b0"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190h-20030916.bin", Description: "China/HK v1.90", Hashes: []string{"05359e5a7c88a69eb25bf5f4b85c3a0b"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190j-20030707.bin", Description: "Japan v1.90", Hashes: []string{"0270d4ca21f18f0f220f131cf9136a14"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190jd-20030707.bin", Description: "Japan v1.90 debug", Hashes: []string{"760e38eb17cded716bc756a8965643be"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190r-20030916.bin", Description: "Russia v1.90", Hashes: []string{"67d613747c39b50f2ce3fbbed7e8ebe4"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0190u-20031017.bin", Description: "USA v1.90", Hashes: []string{"e973c9c5f78aaff915f9fe09a33bce8e"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200a-20040419.bin", Description: "Asia v2.00", Hashes: []string{"6654c39431dfe83692dd26695d836de7"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200a-20040614.bin", Description: "USA v2.00", Hashes: []string{"d333558cc14561c1fdc334c75d5f37b7"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200ad-20040419.bin", Description: "Asia v2.00 debug", Hashes: []string{"8d83489e564f72a5b31bfa56bcf6e108"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200e-20040614.bin", Description: "Europe v2.00", Hashes: []string{"dc752f160044f2ed5fc1f4964db2a095"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200e-20041115.bin", Description: "Europe v2.00", Hashes: []string{"dc996660a859a8c23a952f633a409556"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200h-20040419.bin", Description: "China/HK v2.00", Hashes: []string{"b82f12f8b186cbf8e2e3fdc337044684"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200j-20040323.bin", Description: "Japan v2.00", Hashes: []string{"2253c3d364ff49c234a38f3d631fc31a"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200j-20040614.bin", Description: "Japan v2.00", Hashes: []string{"0eee5d1c779aa50e94edd168b4ebf42e"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0200u-20040419.bin", Description: "USA v2.00", Hashes: []string{"99c5cc568676fa3d4c4d2c4a492d1ec5"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0201a-20040915.bin", Description: "Asia v2.01", Hashes: []string{"d5bb76a9e7c11f59a4263f92f3053c75"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0201h-20040915.bin", Description: "China/HK v2.01", Hashes: []string{"546e7d6f258d2c229812f9158cfc204d"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0201j-20040819.bin", Description: "Japan v2.01", Hashes: []string{"8d6e89165d77a6eca1528b0811aabc7a"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0201u-20040915.bin", Description: "USA v2.01", Hashes: []string{"6bf1fce90e751b6df66aed444057ff94"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0202a-20050121.bin", Description: "Asia v2.02", Hashes: []string{"62fab6930f454c651eecadf42db4927f"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0202e-20050121.bin", Description: "Europe v2.02", Hashes: []string{"b411fffe806a458518a874db389e7271"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0202h-20050121.bin", Description: "China/HK v2.02", Hashes: []string{"f4b982aa5a5c1adf7e9f7dc3cba7e3a4"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0202j-20041222.bin", Description: "Japan v2.02", Hashes: []string{"f854a87b6be3e88841e2de3c8d2d101e"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0202u-20050121.bin", Description: "USA v2.02", Hashes: []string{"7c6358d74fd121ca3df740f9db3195a4"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0203a-20050304.bin", Description: "Asia v2.03", Hashes: []string{"34d8e1f5ba1fa8b41c1ec2ed2b920eac"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0203e-20050304.bin", Description: "Europe v2.03", Hashes: []string{"810f42b2fcf7d95dfdc9e5f629d88cc4"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0205u-20051020.bin", Description: "USA v2.05", Hashes: []string{"11d888baf64c6c4baa2f673f0aa8ce43"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0230a-20080220.bin", Description: "USA v2.30", Hashes: []string{"21038400dc633070a78ad53090c53017"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0230e-20080220.bin", Description: "Europe v2.30", Hashes: []string{"dc69f0643a3030aaa4797501b483d6c4"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0230j-20080220.bin", Description: "Japan v2.30", Hashes: []string{"30d56e79d89fbddf10938fa67fe3f34e"}},
	{Kind: model.ProvisionBIOS, Filename: "ps2-0250e-20100415.bin", Description: "Europe v2.50", Hashes: []string{"93ea3bcee4252627919175ff1b16a1d9"}},
	{Kind: model.ProvisionBIOS, Filename: "scph10000.bin", Description: "Japan v1.00 SCPH-10000", Hashes: []string{"b7fa11e87d51752a98b38e3e691cbf17"}},
	{Kind: model.ProvisionBIOS, Filename: "scph39001.bin", Description: "USA v1.60 SCPH-39001", Hashes: []string{"d5ce2c7d119f563ce04bc04dbc3a323e"}},
	{Kind: model.ProvisionBIOS, Filename: "scph39004.bin", Description: "Europe v1.60 SCPH-39004", Hashes: []string{"1ad977bb539fc9448a08ab276a836bbc"}},
	{Kind: model.ProvisionBIOS, Filename: "scph50000.bin", Description: "Japan v1.70 SCPH-50000", Hashes: []string{"d3f1853a16c2ec18f3cd1ae655213308"}},
	{Kind: model.ProvisionBIOS, Filename: "scph70000.bin", Description: "Japan v2.00 SCPH-70000", Hashes: []string{"0bf988e9c7aaa4c051805b0fa6eb3387"}},
	{Kind: model.ProvisionBIOS, Filename: "scph70004.bin", Description: "Europe v2.00 SCPH-70004", Hashes: []string{"d333558cc14561c1fdc334c75d5f37b7"}},
	{Kind: model.ProvisionBIOS, Filename: "scph70012.bin", Description: "USA v2.00 SCPH-70012", Hashes: []string{"d333558cc14561c1fdc334c75d5f37b7"}},
	{Kind: model.ProvisionBIOS, Filename: "scph77001.bin", Description: "USA v2.20 SCPH-77001", Hashes: []string{"af60e6d1a939019d55e5b330d24b1c25"}},
}
