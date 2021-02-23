package go_uid_generator

/**
 * Represents a unique id generator.
 *
 * @author evan
 */
type UidGenerator interface {

	/**
	Get a unique ID
	*/
	GetUID() (int64, error)

	/**

	 */
	ParseUID(uid int64) string
}
