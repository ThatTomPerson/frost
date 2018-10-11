test:
	mage -compile test-binary
	(cd test; ../test-binary install;)
.PHONY: test