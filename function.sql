CREATE OR REPLACE FUNCTION getDomains()
RETURNS TABLE (
  id uuid,
  name varchar,
  website varchar,
  coordinates geography,
  description varchar,
  rating float8
)
LANGUAGE plpgsql
AS
$$
#variable_conflict use_column
BEGIN
	UPDATE "MY_TABLE" SET website = substring(website from '^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/?\n]+)');
	RETURN QUERY SELECT spot.* FROM "MY_TABLE" spot JOIN(SELECT website, COUNT(website) FROM "MY_TABLE" GROUP BY website HAVING COUNT(website) > 1) duplicate ON spot.website = duplicate.website ORDER BY spot.website;
END;
$$;
SELECT * FROM getDomains()
