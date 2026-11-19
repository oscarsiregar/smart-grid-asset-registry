import styles from "./styles.module.css";

type Properties = {
	assetId: string;
};

/** An asset ID in a small monospace tag. */
const IdTag: React.FC<Properties> = ({ assetId }: Properties) => (
	<span className={styles["tag"]}>{assetId}</span>
);

export { IdTag };
