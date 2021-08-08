package redisearch

import "fmt"

const (
	// FieldTypeText Allows full-text search queries against the value in this field.
	FieldTypeText string = "TEXT"
	// FieldTypeTag Allows exact-match queries, such as categories or primary keys,
	// against the value in this field. For more information, see https://oss.redislabs.com/redisearch/Tags/
	FieldTypeTag string = "TAG"
	// FieldTypeNumeric Allows numeric range queries against the value in this field.
	// See https://oss.redislabs.com/redisearch/Query_Syntax/ for details on how to use numeric ranges.
	FieldTypeNumeric string = "NUMERIC"
	// FieldTypeGeo Allows geographic range queries against the value in this field.
	// The value of the field must be a string containing a longitude (first) and latitude separated by a comma.
	FieldTypeGeo string = "GEO"

	// IndexFlagNoOffsets If set, we do not store term offsets for documents (saves memory, does not allow exact searches or highlighting). Implies NOHL .
	IndexFlagNoOffsets string = "NOOFFSETS"
	// IndexFlagNoHl Conserves storage space and memory by disabling highlighting support.
	// If set, we do not store corresponding byte offsets for term positions. NOHL is also implied by NOOFFSETS
	IndexFlagNoHl string = "NOHL"
	// IndexFlagNoFields If set, we do not store field bits for each term. Saves memory, does not allow filtering by specific fields.
	IndexFlagNoFields string = "NOFIELDS"
	// IndexFlagNoFreqs If set, we avoid saving the term frequencies in the index. This saves memory but does not allow sorting based on the frequencies of a given term within the document.
	IndexFlagNoFreqs string = "NOFREQS"
	// IndexFlagSkipInitialScan If set, we do not scan and index.
	IndexFlagSkipInitialScan string = "SKIPINITIALSCAN"
	// IndexFlagMaxTextFields For efficiency, RediSearch encodes indexes differently if they are created with less than 32 text fields.
	// This option forces RediSearch to encode indexes as if there were more than 32 text fields, which allows you to add additional fields (beyond 32) using FT.ALTER .
	IndexFlagMaxTextFields string = "MAXTEXTFIELDS"
	// SearchFlagVerbatim if set, we do not try to use stemming for query expansion but search the query terms verbatim.
	SearchFlagVerbatim = "VERBATIM"
	// SearchFlagNoStopWords If set, we do not filter stopwords from the query.
	SearchFlagNoStopWords = "NOSTOPWORDS"

	// NOT SUPPORTED ATM, search results parser needs to be updated to understand these
	// SearchFlagNoContent If it appears after the query, we only return the document ids and not the content. This is useful if RediSearch is only an index on an external document collection
	// SearchFlagNoContent = "NOCONTENT"
	// SearchFlagWithScores If set, we also return the relative internal score of each document. this can be used to merge results from multiple instances
	// SearchFlagWithScores = "WITHSCORES"
	// SearchFlagWithPayloads //  If set, we retrieve optional document payloads (see FT.ADD).
	// the payloads follow the document id, and if WITHSCORES was set, follow the scores.
	// SearchFlagWithPayloads = "WITHPAYLOADS"
	// SearchFlagWithSortKeys Only relevant in conjunction with SORTBY . Returns the value of the sorting key, right after the id and score and /or payload if requested.
	// This is usually not needed by users, and exists for distributed search coordination purposes.
	// SearchFlagWithSortKeys = "WITHSORTKEYS"
)

type SchemaOpt []interface{}

// SchemaOptNoStem Text fields can have the NOSTEM argument which will disable stemming when indexing its values.
// This may be ideal for things like proper names.
func SchemaOptNoStem() SchemaOpt {
	return []interface{}{"NOSTEM"}
}

// SchemaOptWeight For TEXT fields, declares the importance of this field when calculating result accuracy.
// This is a multiplication factor, and defaults to 1 if not specified.
func SchemaOptWeight(weight float32) SchemaOpt {
	return []interface{}{"WEIGHT", fmt.Sprintf("%.1f", weight)}
}

// SchemaOptSortable Numeric, tag or text fields can have the optional SORTABLE argument that allows the user to later sort the results
// by the value of this field (this adds memory overhead so do not declare it on large text fields).
func SchemaOptSortable() SchemaOpt {
	return []interface{}{"SORTABLE"}
}

// SchemaOptTagSeparator or TAG fields, indicates how the text contained in the field is to be split into individual tags.
// The default is , . The value must be a single character
func SchemaOptTagSeparator(character byte) SchemaOpt {
	return []interface{}{"SEPARATOR", character}
}

// SchemaOptNoIndex Fields can have the NOINDEX option, which means they will not be indexed. This is useful in conjunction with SORTABLE,
//	to create fields whose update using PARTIAL will not cause full reindexing of the document.
//	If a field has NOINDEX and doesn't have SORTABLE, it will just be ignored by the index.
func SchemaOptNoIndex() SchemaOpt {
	return []interface{}{"NOINDEX"}
}

// SchemaOptPhonetic Declaring a text field as PHONETIC will perform phonetic matching on it in searches by default.
// The obligatory {matcher} argument specifies the phonetic algorithm and language used. The following matchers are supported:
// * dm:en - Double Metaphone for English
// * dm:fr - Double Metaphone for French
// * dm:pt - Double Metaphone for Portuguese
// * dm:es - Double Metaphone for Spanish
// For more details see Phonetic Matching .
func SchemaOptPhonetic(matcher string) SchemaOpt {
	return []interface{}{"PHONETIC", matcher}
}

type FieldSchema struct {
	// Field types can be numeric, textual or geographical.
	// See FieldDataType constants
	Type    string
	Options []SchemaOpt
}
type IndexOptions struct {
	// The index name to create. If it exists the old spec will be overwritten
	// This is a REQUIRED field
	IndexName string
	// Tells the index which keys it should index.
	// You can add several prefixes to index. Since the argument is optional, the default is * (all keys)
	Prefix []string
	// {filter} is a filter expression with the full RediSearch aggregation expression language.
	// It is possible to use @__key to access the key that was just added/changed.
	// A field can be used to set field name by passing 'FILTER @indexName=="myindexname"'
	Filter string
	// If set indicates the default language for documents in the index. Default to English.
	// The supported languages are: Arabic, Basque, Catalan, Danish, Dutch, English, Finnish, French, German, Greek, Hungarian,
	// Indonesian, Irish, Italian, Lithuanian, Nepali, Norwegian, Portuguese, Romanian, Russian, Spanish, Swedish, Tamil, Turkish, Chinese
	Language string
	// If set indicates the document field that should be used as the document language.
	LanguageField string
	// If set indicates the default score for documents in the index. Default score is 1.0.
	Score float32
	// If set indicates the document field that should be used as the document's rank based on the user's ranking.
	// Ranking must be between 0.0 and 1.0. If not set the default score is 1.
	ScoreField string
	// If set indicates the document field that should be used as a binary safe payload string to the document,
	// that can be evaluated at query time by a custom scoring function, or retrieved to the client.
	PayloadField string
	// Index options/flags
	Flags []string
	// If set, we set the index with a custom stopword list, to be ignored during indexing and search time.
	// If not set, we take the default list of stopwords.
	StopWords []string
	// Create a lightweight temporary index which will expire after the specified period of inactivity.
	// The internal idle timer is reset whenever the index is searched or added to. Because such indexes are lightweight,
	// you can create thousands of such indexes without negative performance implications and therefore you should consider using SKIPINITIALSCAN to avoid costly scanning.
	Temporary float32
	// After the SCHEMA keyword we define the index fields.
	// The field name is the name of the field within the hashes that this index follows. Field types can be numeric, textual or geographical.
	Schema map[string]FieldSchema
}

type FieldFilter struct {
	NumericFieldName string
	Min              float32
	Max              float32
	Exclusive        bool // Do not include Min and Max, only in between
}

type GeoFilter struct {
	GeoFieldName string
	Latitude     float32
	Longitude    float32
	Radius       float32
	// valid values are m|km|mi|ft
	// by default set to 'm'
	Unit string
}

type Highlight struct {
	// If present, must be the first argument. This should be followed by the number of fields to highlight,
	// which itself is followed by a list of fields. Each field present is highlighted.
	// If no FIELDS directive is passed, then all fields returned are highlighted.
	Fields   []string
	OpenTag  string
	CloseTag string
}

type Summarize struct {
	// If present, must be the first argument. This should be followed by the number of fields to summarize,
	// which itself is followed by a list of fields. Each field present is summarized.
	//If no FIELDS directive is passed, then all fields returned are summarized.
	Fields []string
	// How many fragments should be returned. If not specified, a default of 3 is used.
	Fragments int
	// The number of context words each fragment should contain. Context words surround the found term.
	// A higher value will return a larger block of text. If not specified, the default value is 20.
	Length int
	// The string used to divide between individual summary snippets.
	// The default is ... which is common among search engines; but you may override this with any other string if you desire to programmatically divide them later on.
	// You may use a newline sequence, as newlines are stripped from the result body anyway (thus, it will not be conflated with an embedded newline in the text)
	Separator string
}

type SortBy struct {
	FieldName  string
	Descending bool
}

type Limit struct {
	Offset int
	Max    int
}

type SearchOptions struct {
	// IndexName The index name. The index must be first created with FT.CREATE .
	IndexName string
	// Query the text query to search. If it's more than a single word, put it in quotes. Refer to query syntax for more details.
	// See https://oss.redislabs.com/redisearch/Query_Syntax/
	Query string
	// Filters numeric_field min max : If set, and numeric_field is defined as a numeric field in FT.CREATE,
	// we will limit results to those having numeric values ranging between min and max
	// min and max follow ZRANGE syntax, and can be -inf , +inf and use ( for exclusive ranges
	Filters []FieldFilter
	// GeoFilter If set, we filter the results to a given radius from lon and lat.
	// Radius is given as a number and units. See GEORADIUS for more details.
	GeoFilter *GeoFilter
	// InKeys If set, we limit the result to a given set of keys specified in the list.
	// Non-existent keys are ignored - unless all the keys are non-existent.
	InKeys []string
	// InFields If set, filter the results to ones appearing only in specific fields of the document, like title or URL.
	InFields []string
	// Return Use this keyword to limit which fields from the document are returned.
	Return []string
	// Summarize Use this option to return only the sections of the field which contain the matched text.
	// See https://oss.redislabs.com/redisearch/Highlight/ for more details
	Summarize *Summarize
	// Highlight Use this option to format occurrences of matched text.
	// See https://oss.redislabs.com/redisearch/Highlight/ for more details
	Highlight *Highlight
	// Slop If set, we allow a maximum of N intervening number of unmatched offsets between phrase terms. (i.e the slop for exact phrases is 0)
	Slop *int
	// Language If set, we use a stemmer for the supplied language during search for query expansion.
	//If querying documents in Chinese, this should be set to chinese in order to properly tokenize the query terms.
	//Defaults to English. If an unsupported language is sent, the command returns an error. See FT.ADD for the list of languages.
	Language string
	// Expander If set, we will use a custom query expander instead of the stemmer.
	// See https://oss.redislabs.com/redisearch/Extensions/
	Expander string
	// Scorer If set, we will use a custom scoring function defined by the user.
	// See https://oss.redislabs.com/redisearch/Extensions/
	Scorer string
	// Payload Add an arbitrary, binary safe payload that will be exposed to custom scoring functions.
	// See https://oss.redislabs.com/redisearch/Extensions/
	Payload string
	// SortBy {field} [ASC|DESC] : If specified, the results are ordered by the value of this field. This applies to both text and numeric fields.
	SortBy *SortBy
	// Limit first num : Limit the results to the offset and number of results given. Note that the offset is zero-indexed.
	// The default is 0 10, which returns 10 items starting from the first result.
	Limit *Limit
	// Flags see SearchFlag* constants
	Flags []string
}
