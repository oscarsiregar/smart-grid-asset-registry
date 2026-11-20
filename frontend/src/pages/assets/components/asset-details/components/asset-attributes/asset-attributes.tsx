import { Link } from "~/libs/components/link/link.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { type AssetDto, type AssetNodeDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { getAssetPath } from "../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

type Properties = {
	asset: AssetDto;
	/** The parent from the path; `null` for a substation. */
	parent: AssetNodeDto | null;
};

const getText = (value: null | string): string =>
	value ?? AssetsPageMessage.NOT_RECORDED;

// Numbers are shown by value with their unit (spec FR-010).
const getMeasurement = (value: null | number, template: string): string =>
	value === null
		? AssetsPageMessage.NOT_RECORDED
		: configureString(template, { value: String(value) });

/**
 * The attributes the intro does not show — parent first, then the measured
 * and recorded values; a blank one says "Not recorded" (spec FR-008, FR-009,
 * spec 005 FR-012). Asset ID, type, name and status are in the intro.
 */
const AssetAttributes: React.FC<Properties> = ({
	asset,
	parent,
}: Properties) => {
	const attributes = [
		{
			label: AssetsPageMessage.ATTRIBUTE_VOLTAGE,
			value: getMeasurement(asset.voltageKv, AssetsPageMessage.UNIT_KV),
		},
		{
			label: AssetsPageMessage.ATTRIBUTE_RATING,
			value: getMeasurement(asset.ratingKva, AssetsPageMessage.UNIT_KVA),
		},
		{
			label: AssetsPageMessage.ATTRIBUTE_COMMISSIONED_DATE,
			value: getText(asset.commissionedDate),
		},
		{
			label: AssetsPageMessage.ATTRIBUTE_MANUFACTURER,
			value: getText(asset.manufacturer),
		},
		{ label: AssetsPageMessage.ATTRIBUTE_MODEL, value: getText(asset.model) },
		{
			label: AssetsPageMessage.ATTRIBUTE_SERIAL_NUMBER,
			value: getText(asset.serialNumber),
		},
	];

	return (
		<dl className={styles["attributes"]}>
			{asset.parentAssetId !== null && (
				<div className={styles["attribute"]}>
					<dt className={styles["term"]}>
						{AssetsPageMessage.ATTRIBUTE_PARENT}
					</dt>
					<dd className={styles["value"]}>
						<Link
							className={styles["link"]}
							to={getAssetPath(asset.parentAssetId)}
						>
							<span className={styles["id"]}>{asset.parentAssetId}</span>
							{parent && <span>{parent.assetName}</span>}
						</Link>
					</dd>
				</div>
			)}
			{attributes.map(({ label, value }) => (
				<div className={styles["attribute"]} key={label}>
					<dt className={styles["term"]}>{label}</dt>
					<dd className={styles["value"]}>{value}</dd>
				</div>
			))}
		</dl>
	);
};

export { AssetAttributes };
