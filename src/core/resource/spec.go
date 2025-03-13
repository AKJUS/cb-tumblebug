/*
Copyright 2019 The Cloud-Barista Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package resource is to manage multi-cloud infra resource
package resource

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cb-tumblebug/src/core/common"
	clientManager "github.com/cloud-barista/cb-tumblebug/src/core/common/client"
	"github.com/cloud-barista/cb-tumblebug/src/core/model"
	validator "github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	//"github.com/cloud-barista/cb-tumblebug/src/core/mci"

	_ "github.com/go-sql-driver/mysql"
)

// TbSpecReqStructLevelValidation is a function to validate 'TbSpecReq' object.
func TbSpecReqStructLevelValidation(sl validator.StructLevel) {

	u := sl.Current().Interface().(model.TbSpecReq)

	err := common.CheckString(u.Name)
	if err != nil {
		// ReportError(field interface{}, fieldName, structFieldName, tag, param string)
		sl.ReportError(u.Name, "name", "Name", err.Error(), "")
	}
}

// ConvertSpiderSpecToTumblebugSpec accepts an Spider spec object, converts to and returns an TB spec object
func ConvertSpiderSpecToTumblebugSpec(spiderSpec model.SpiderSpecInfo) (model.TbSpecInfo, error) {
	if spiderSpec.Name == "" {
		err := fmt.Errorf("convertSpiderSpecToTumblebugSpec failed; spiderSpec.Name == \"\" ")
		emptyTumblebugSpec := model.TbSpecInfo{}
		return emptyTumblebugSpec, err
	}

	tumblebugSpec := model.TbSpecInfo{}

	tumblebugSpec.Name = spiderSpec.Name
	tumblebugSpec.CspSpecName = spiderSpec.Name
	tumblebugSpec.RegionName = spiderSpec.Region

	tempUint64, _ := strconv.ParseUint(spiderSpec.VCpu.Count, 10, 16)
	tumblebugSpec.VCPU = uint16(tempUint64)
	tempFloat64, _ := strconv.ParseFloat(spiderSpec.MemSizeMiB, 32)
	tumblebugSpec.MemoryGiB = float32(tempFloat64 / 1024)
	tempFloat64, _ = strconv.ParseFloat(spiderSpec.DiskSizeGB, 32)
	tumblebugSpec.DiskSizeGB = float32(tempFloat64)

	// GPU(Accelerator) information conversion
	if len(spiderSpec.Gpu) > 0 {
		// Set AcceleratorType to "gpu" when GPU exists
		tumblebugSpec.AcceleratorType = "gpu"

		// Use the first GPU information
		firstGpu := spiderSpec.Gpu[0]

		// Combine Mfr and Model to form AcceleratorModel
		if firstGpu.Mfr != "" && firstGpu.Model != "" {
			// Check if Model already starts with Mfr to avoid duplication
			if strings.HasPrefix(firstGpu.Model, firstGpu.Mfr) {
				// Model already includes Mfr, so just use Model
				tumblebugSpec.AcceleratorModel = firstGpu.Model
			} else {
				// Model doesn't include Mfr, so combine them
				tumblebugSpec.AcceleratorModel = firstGpu.Mfr + " " + firstGpu.Model
			}
		} else if firstGpu.Model != "" {
			tumblebugSpec.AcceleratorModel = firstGpu.Model
		} else if firstGpu.Mfr != "" {
			tumblebugSpec.AcceleratorModel = firstGpu.Mfr
		}

		// Convert GPU count
		if firstGpu.Count != "" && firstGpu.Count != "-1" {
			tempCount, err := strconv.ParseUint(firstGpu.Count, 10, 8)
			if err == nil {
				tumblebugSpec.AcceleratorCount = uint8(tempCount)
			}
		}

		// Convert GPU memory size
		if firstGpu.MemSizeGB != "" && firstGpu.MemSizeGB != "-1" {
			tempMemory, err := strconv.ParseFloat(firstGpu.MemSizeGB, 32)
			if err == nil {
				tumblebugSpec.AcceleratorMemoryGB = float32(tempMemory)
			}
		}

		// Log if there are multiple GPUs defined
		if len(spiderSpec.Gpu) > 1 {
			log.Warn().Msgf("Spec %s has multiple GPUs defined (%d GPUs). Only using the first GPU information.",
				spiderSpec.Name, len(spiderSpec.Gpu))
		}
	}

	return tumblebugSpec, nil
}

// LookupSpecList accepts Spider conn config,
// lookups and returns the list of all specs in the region of conn config
// in the form of the list of Spider spec objects
func LookupSpecList(connConfig string) (model.SpiderSpecList, error) {

	if connConfig == "" {
		content := model.SpiderSpecList{}
		err := fmt.Errorf("LookupSpec called with empty connConfig.")
		log.Error().Err(err).Msg("")
		return content, err
	}

	var callResult model.SpiderSpecList
	client := resty.New()
	client.SetTimeout(10 * time.Minute)
	url := model.SpiderRestUrl + "/vmspec"
	method := "GET"
	requestBody := model.SpiderConnectionName{}
	requestBody.ConnectionName = connConfig

	err := clientManager.ExecuteHttpRequest(
		client,
		method,
		url,
		nil,
		clientManager.SetUseBody(requestBody),
		&requestBody,
		&callResult,
		clientManager.MediumDuration,
	)

	if err != nil {
		log.Trace().Err(err).Msg("")
		content := model.SpiderSpecList{}
		return content, err
	}

	temp := callResult
	return temp, nil

}

// LookupSpec accepts Spider conn config and CSP spec name, lookups and returns the Spider spec object
func LookupSpec(connConfig string, specName string) (model.SpiderSpecInfo, error) {

	if connConfig == "" {
		content := model.SpiderSpecInfo{}
		err := fmt.Errorf("LookupSpec() called with empty connConfig.")
		log.Error().Err(err).Msg("")
		return content, err
	} else if specName == "" {
		content := model.SpiderSpecInfo{}
		err := fmt.Errorf("LookupSpec() called with empty specName.")
		log.Error().Err(err).Msg("")
		return content, err
	}

	client := resty.New()
	client.SetTimeout(2 * time.Minute)
	url := model.SpiderRestUrl + "/vmspec/" + specName
	method := "GET"
	requestBody := model.SpiderConnectionName{}
	requestBody.ConnectionName = connConfig
	callResult := model.SpiderSpecInfo{}

	err := clientManager.ExecuteHttpRequest(
		client,
		method,
		url,
		nil,
		clientManager.SetUseBody(requestBody),
		&requestBody,
		&callResult,
		clientManager.MediumDuration,
	)

	if err != nil {
		log.Error().Err(err).Msg("")
		return callResult, err
	}

	return callResult, nil
}

// FetchSpecsForConnConfig lookups all specs for region of conn config, and saves into TB spec objects
func FetchSpecsForConnConfig(connConfigName string, nsId string) (uint, error) {
	log.Debug().Msg("FetchSpecsForConnConfig(" + connConfigName + ")")
	specCount := uint(0)

	connConfig, err := common.GetConnConfig(connConfigName)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot GetConnConfig in %s", connConfigName)
		return specCount, err
	}

	specsInConnection, err := LookupSpecList(connConfigName)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot LookupSpecList in %s", connConfigName)
		return specCount, err
	}

	for _, spec := range specsInConnection.Vmspec {
		spiderSpec := spec
		//log.Info().Msgf("Found spec in the map: %s", spiderSpec.Name)
		tumblebugSpec, errConvert := ConvertSpiderSpecToTumblebugSpec(spiderSpec)
		if errConvert != nil {
			log.Error().Err(errConvert).Msg("Cannot ConvertSpiderSpecToTumblebugSpec")
		} else {
			key := GetProviderRegionZoneResourceKey(connConfig.ProviderName, connConfig.RegionDetail.RegionName, "", spec.Name)
			tumblebugSpec.Name = key
			tumblebugSpec.ConnectionName = connConfig.ConfigName
			tumblebugSpec.ProviderName = strings.ToLower(connConfig.ProviderName)
			tumblebugSpec.RegionName = connConfig.RegionDetail.RegionName
			tumblebugSpec.InfraType = "vm" // default value
			tumblebugSpec.SystemLabel = "auto-gen"
			tumblebugSpec.CostPerHour = 99999999.9
			tumblebugSpec.EvaluationScore01 = -99.9

			_, err := RegisterSpecWithInfo(nsId, &tumblebugSpec, true)
			if err != nil {
				log.Error().Err(err).Msg("")
				return 0, err
			}
			specCount++
		}

	}
	return specCount, nil
}

// FetchSpecsForAllConnConfigs gets all conn configs from Spider, lookups all specs for each region of conn config, and saves into TB spec objects
func FetchSpecsForAllConnConfigs(nsId string) (connConfigCount uint, specCount uint, err error) {

	err = common.CheckString(nsId)
	if err != nil {
		log.Error().Err(err).Msg("")
		return 0, 0, err
	}

	connConfigs, err := common.GetConnConfigList(model.DefaultCredentialHolder, true, true)
	if err != nil {
		log.Error().Err(err).Msg("")
		return 0, 0, err
	}

	for _, connConfig := range connConfigs.Connectionconfig {
		temp, _ := FetchSpecsForConnConfig(connConfig.ConfigName, nsId)
		specCount += temp
		connConfigCount++
	}
	return connConfigCount, specCount, nil
}

// RegisterSpecWithCspResourceId accepts spec creation request, creates and returns an TB spec object
func RegisterSpecWithCspResourceId(nsId string, u *model.TbSpecReq, update bool) (model.TbSpecInfo, error) {

	content := model.TbSpecInfo{}

	err := common.CheckString(nsId)
	if err != nil {
		log.Error().Err(err).Msg("")
		return content, err
	}

	res, err := LookupSpec(u.ConnectionName, u.CspSpecName)
	if err != nil {
		log.Error().Err(err).Msgf("cannot LookupSpec ConnectionName(%s), CspResourceId(%s)", u.ConnectionName, u.CspSpecName)
		return content, err
	}

	content, err = ConvertSpiderSpecToTumblebugSpec(res)
	if err != nil {
		log.Error().Err(err).Msg("cannot RegisterSpecWithCspResourceId")
		return content, err
	}

	content.Namespace = nsId
	content.Id = u.Name
	content.Name = u.Name
	content.ConnectionName = u.ConnectionName
	content.AssociatedObjectList = []string{}

	// "INSERT INTO `spec`(`namespace`, `id`, ...) VALUES ('nsId', 'content.Id', ...);
	_, err = model.ORM.Insert(&content)
	if err != nil {
		log.Error().Err(err).Msg("Cannot insert data to RDB")
	} else {
		log.Trace().Msg("SQL: Insert success")
	}

	return content, nil
}

// RegisterSpecWithInfo accepts spec creation request, creates and returns an TB spec object
func RegisterSpecWithInfo(nsId string, content *model.TbSpecInfo, update bool) (model.TbSpecInfo, error) {

	err := common.CheckString(nsId)
	if err != nil {
		temp := model.TbSpecInfo{}
		log.Error().Err(err).Msg("")
		return temp, err
	}

	content.Namespace = nsId
	content.Id = content.Name
	content.AssociatedObjectList = []string{}

	// "INSERT INTO `spec`(`namespace`, `id`, ...) VALUES ('nsId', 'content.Id', ...);
	// Attempt to insert the new record
	_, err = model.ORM.Insert(content)
	if err != nil {
		if update {
			// If insert fails and update is true, attempt to update the existing record
			_, updateErr := model.ORM.Update(content, &model.TbSpecInfo{Namespace: content.Namespace, Id: content.Id})
			if updateErr != nil {
				log.Error().Err(updateErr).Msg("Error updating spec after insert failure")
				return *content, updateErr
			} else {
				log.Trace().Msg("SQL: Update success after insert failure")
			}
		} else {
			log.Error().Err(err).Msg("Error inserting spec and update flag is false")
			return *content, err
		}
	} else {
		log.Trace().Msg("SQL: Insert success")
	}

	return *content, nil
}

// RegisterSpecWithInfoInBulk register a list of specs in bulk
func RegisterSpecWithInfoInBulk(specList []model.TbSpecInfo) error {
	// Insert in bulk
	// batch size is 90 due to the limit of SQL
	batchSize := 90

	total := len(specList)
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := specList[i:end]

		session := model.ORM.NewSession()
		defer session.Close()
		if err := session.Begin(); err != nil {
			log.Error().Err(err).Msg("Failed to begin transaction")
			return err
		}

		affected, err := session.Insert(&batch)
		if err != nil {
			session.Rollback()
			log.Error().Err(err).Msg("Error inserting specs in bulk")
			return err
		} else {
			if err := session.Commit(); err != nil {
				log.Error().Err(err).Msg("Failed to commit transaction")
				return err
			}
			log.Trace().Msgf("Bulk insert success: %d records affected", affected)
		}
	}
	return nil
}

// RemoveDuplicateSpecsInSQL is to remove duplicate specs in db to refine batch insert duplicates
func RemoveDuplicateSpecsInSQL() error {
	sqlStr := `
	DELETE FROM TbSpecInfo
	WHERE rowid NOT IN (
		SELECT MAX(rowid)
		FROM TbSpecInfo
		GROUP BY Namespace, Id
	);
	`
	_, err := model.ORM.Exec(sqlStr)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting duplicate specs")
		return err
	}
	log.Info().Msg("Duplicate specs removed successfully")

	return nil
}

// Range struct is for 'FilterSpecsByRange'
type Range struct {
	Min float32 `json:"min"`
	Max float32 `json:"max"`
}

// GetSpec accepts namespace Id and specKey(Id,CspResourceName,...), and returns the TB spec object
func GetSpec(nsId string, specKey string) (model.TbSpecInfo, error) {
	if err := common.CheckString(nsId); err != nil {
		log.Error().Err(err).Msg("Invalid namespace ID")
		return model.TbSpecInfo{}, err
	}

	log.Debug().Msg("[Get spec] " + specKey)

	// make comparison case-insensitive
	nsId = strings.ToLower(nsId)
	specKey = strings.ToLower(specKey)

	// ex: tencent+ap-jakarta+ubuntu22.04
	spec := model.TbSpecInfo{Namespace: nsId, Id: specKey}
	has, err := model.ORM.Where("LOWER(Namespace) = ? AND LOWER(Id) = ?", nsId, specKey).Get(&spec)
	if err != nil {
		log.Info().Err(err).Msgf("Failed to get spec %s by ID", specKey)
	}
	if has {
		return spec, nil
	}

	// ex: img-487zeit5
	spec = model.TbSpecInfo{Namespace: nsId, CspSpecName: specKey}
	has, err = model.ORM.Where("LOWER(Namespace) = ? AND LOWER(CspSpecName) = ?", nsId, specKey).Get(&spec)
	if err != nil {
		log.Info().Err(err).Msgf("Failed to get spec %s by CspSpecName", specKey)
	}
	if has {
		return spec, nil
	}

	return model.TbSpecInfo{}, fmt.Errorf("The specKey %s not found by any of ID, CspSpecName", specKey)
}

// FilterSpecsByRange accepts criteria ranges for filtering, and returns the list of filtered TB spec objects
func FilterSpecsByRange(nsId string, filter model.FilterSpecsByRangeRequest) ([]model.TbSpecInfo, error) {
	if err := common.CheckString(nsId); err != nil {
		log.Error().Err(err).Msg("Invalid namespace ID")
		return nil, err
	}

	// Start building the query using field names as database column names
	session := model.ORM.Where("Namespace = ?", nsId)

	// Use reflection to iterate over filter struct
	val := reflect.ValueOf(filter)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// Convert the first letter of the field name to lowercase to match typical database column naming conventions
		dbFieldName := strings.ToLower(field.Name[:1]) + field.Name[1:]
		//log.Debug().Msgf("Field: %s, Value: %v", dbFieldName, value)

		if value.Kind() == reflect.Struct {
			// Handle range filters like VCPU, MemoryGiB, etc.
			min := value.FieldByName("Min")
			max := value.FieldByName("Max")

			if min.IsValid() && !min.IsZero() {
				session = session.And(dbFieldName+" >= ?", min.Interface())
			}
			if max.IsValid() && !max.IsZero() {
				session = session.And(dbFieldName+" <= ?", max.Interface())
			}
		} else if value.IsValid() && !value.IsZero() {
			switch value.Kind() {
			case reflect.String:
				cleanValue := ToNamingRuleCompatible(value.String())
				session = session.And(dbFieldName+" LIKE ?", "%"+cleanValue+"%")
				log.Info().Msgf("Filtering by %s: %s", dbFieldName, cleanValue)
			}
		}
	}

	startTime := time.Now()

	var specs []model.TbSpecInfo
	err := session.Find(&specs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute query")
		return nil, err
	}

	elapsedTime := time.Since(startTime)
	log.Info().
		Dur("elapsedTime", elapsedTime).
		Msg("ORM:session.Find(&specs)")

	return specs, nil
}

// UpdateSpec accepts to-be TB spec objects,
// updates and returns the updated TB spec objects
func UpdateSpec(nsId string, specId string, fieldsToUpdate model.TbSpecInfo) (model.TbSpecInfo, error) {

	_, err := model.ORM.Update(&fieldsToUpdate, &model.TbSpecInfo{Namespace: nsId, Id: specId})
	if err != nil {
		log.Error().Err(err).Msg("")
		return fieldsToUpdate, err
	} else {
		log.Trace().Msg("SQL: Update success")
	}

	return fieldsToUpdate, nil
}
